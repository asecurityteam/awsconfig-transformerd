package v1

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type ec2Configuration struct {
	InstanceID string `json:"instanceId"`
	State      struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"state"`
	StateTransitionReason string             `json:"stateTransitionReason"`
	InstanceType          string             `json:"instanceType"`
	LaunchTime            time.Time          `json:"launchTime"`
	NetworkInterfaces     []networkInterface `json:"networkInterfaces"`
	Tags                  []tag              `json:"tags"`
}

type tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type networkInterface struct {
	NetworkInterfaceID string `json:"networkInterfaceId"`
	SubnetID           string `json:"subnetId"`
	VpcID              string `json:"vpcId"`
	Description        string `json:"description"`
	OwnerID            string `json:"ownerId"`
	PrivateDNSName     string `json:"privateDnsName"`
	Association        struct {
		PublicIP      string `json:"publicIp"`
		PublicDNSName string `json:"publicDnsName"`
		IPOwnerID     string `json:"ipOwnerId"`
	} `json:"association"`
	PrivateIPAddresses []struct {
		PrivateIPAddress string `json:"privateIpAddress"`
		PrivateDNSName   string `json:"privateDnsName"`
		Primary          bool   `json:"primary"`
		Association      struct {
			PublicIP      string `json:"publicIp"`
			PublicDNSName string `json:"publicDnsName"`
			IPOwnerID     string `json:"ipOwnerId"`
		} `json:"association"`
	} `json:"privateIpAddresses"`
	Attachment struct {
		AttachTime string `json:"attachTime"`
	} `json:"attachment"`
}

type networkInterfaceDiff struct {
	PreviousValue *networkInterface `json:"previousValue"`
	UpdatedValue  *networkInterface `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type ec2ConfigurationDiff struct {
	PreviousValue *ec2Configuration `json:"previousValue"`
	UpdatedValue  *ec2Configuration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type ec2ChangedPropsARN struct {
	PreviousValue string `json:"previousValue"`
	UpdatedValue  string `json:"updatedValue"`
	ChangeType    string `json:"changeType"`
}

type ec2Transformer struct{}

func (t ec2Transformer) Create(event awsConfigEvent) ([]Output, error) {
	outputs := []Output{}
	baseOutput, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}
	// if a resource is created for the first time, there is no diff.
	// just read the configuration
	var config ec2Configuration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return []Output{}, err
	}

	hasEni := false
	for _, ni := range config.NetworkInterfaces {
		hasEni = true
		output := baseOutput
		change := Change{}
		output.ChangeTime = ni.Attachment.AttachTime
		private, public, dns := extractNetworkInterfaceInfo(&ni)
		change.PrivateIPAddresses = append(change.PrivateIPAddresses, private...)
		change.PublicIPAddresses = append(change.PublicIPAddresses, public...)
		change.Hostnames = append(change.Hostnames, dns...)
		change.ChangeType = added
		output.Changes = append(output.Changes, change)
		outputs = append(outputs, output)
	}
	if !hasEni {
		return []Output{}, nil // Reject EC2s with no ENI
	}

	return outputs, nil
}

func (t ec2Transformer) Update(event awsConfigEvent) ([]Output, error) {
	outputs := []Output{}
	baseOutput, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}

	// addedChange := Change{ChangeType: added}
	// deletedChange := Change{ChangeType: deleted}
	// If an update was detected, check to see if any changes to the NetworkInterfaces occurred
	for k, v := range event.ConfigurationItemDiff.ChangedProperties {
		if !strings.HasPrefix(k, "Configuration.NetworkInterfaces.") {
			continue
		}
		var diff networkInterfaceDiff
		if err := json.Unmarshal(v, &diff); err != nil {
			return []Output{}, err
		}
		output := baseOutput
		ni := diff.UpdatedValue
		change := Change{}
		// changes := &addedChange
		if diff.ChangeType == delete {
			ni = diff.PreviousValue
			// changes = &deletedChange
		} else {
			output.ChangeTime = ni.Attachment.AttachTime
		}
		private, public, dns := extractNetworkInterfaceInfo(ni)
		change.PrivateIPAddresses = append(change.PrivateIPAddresses, private...)
		change.PublicIPAddresses = append(change.PublicIPAddresses, public...)
		change.Hostnames = append(change.Hostnames, dns...)
		outputs = append(outputs, output)
	}

	// TODO: do we have to remove (not send) add ENI events that also show up as delete events?

	// We need to compute the symmetric difference of the added changes and the removed changes
	// i.e. remove entries that show up as both added and removed
	// symmetricDifference(&addedChange, &deletedChange)
	// if len(addedChange.PrivateIPAddresses) > 0 || len(addedChange.PublicIPAddresses) > 0 || len(addedChange.Hostnames) > 0 {
	// 	output.Changes = append(output.Changes, addedChange)
	// }
	// if len(deletedChange.PrivateIPAddresses) > 0 || len(deletedChange.PublicIPAddresses) > 0 || len(deletedChange.Hostnames) > 0 {
	// 	output.Changes = append(output.Changes, deletedChange)
	// }
	return outputs, nil
}

func (t ec2Transformer) Delete(event awsConfigEvent) ([]Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}

	// if a resource is deleted, the tags are no longer present in the base object.
	// we must fetch them from the previous configuration.
	changeProps := event.ConfigurationItemDiff.ChangedProperties

	if output.ARN == "" {
		previousARNRaw, ok := changeProps["ARN"]
		if !ok {
			return []Output{}, ErrMissingValue{Field: "ARN"}
		}
		var changedPropsARN ec2ChangedPropsARN
		if err := json.Unmarshal(previousARNRaw, &changedPropsARN); err != nil {
			return []Output{}, err
		}
		output.ARN = changedPropsARN.PreviousValue
	}

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return []Output{}, errors.New("Invalid configuration diff")
	}
	var configDiff ec2ConfigurationDiff
	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return []Output{}, err
	}
	for _, tag := range configDiff.PreviousValue.Tags {
		output.Tags[tag.Key] = tag.Value
	}

	// fetch network information from the previous configuration
	change := extractEC2NetworkInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)
	return []Output{output}, nil
}

// remove changes that show up as both added and deleted
func symmetricDifference(a, b *Change) {
	aPrivate := a.PrivateIPAddresses
	a.PrivateIPAddresses = sliceDiff(aPrivate, b.PrivateIPAddresses)
	b.PrivateIPAddresses = sliceDiff(b.PrivateIPAddresses, aPrivate)

	aPublic := a.PublicIPAddresses
	a.PublicIPAddresses = sliceDiff(aPublic, b.PublicIPAddresses)
	b.PublicIPAddresses = sliceDiff(b.PublicIPAddresses, aPublic)

	aHostnames := a.Hostnames
	a.Hostnames = sliceDiff(aHostnames, b.Hostnames)
	b.Hostnames = sliceDiff(b.Hostnames, aHostnames)
}

func sliceDiff(a, b []string) []string {
	m := make(map[string]bool)
	diff := []string{}
	for _, v := range b {
		m[v] = true
	}

	for _, v := range a {
		if _, ok := m[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

// extract network interface information from an ec2 configuration
func extractEC2NetworkInfo(config *ec2Configuration) Change {
	change := Change{}
	for i := range config.NetworkInterfaces {
		private, public, dns := extractNetworkInterfaceInfo(&config.NetworkInterfaces[i])
		change.PrivateIPAddresses = append(change.PrivateIPAddresses, private...)
		change.PublicIPAddresses = append(change.PublicIPAddresses, public...)
		change.Hostnames = append(change.Hostnames, dns...)
	}
	return change
}

// extracts privateIPAddresses, publicIPAddresses, and public DNS names
func extractNetworkInterfaceInfo(ni *networkInterface) ([]string, []string, []string) {
	privateIPAddresses := []string{}
	publicIPAddresses := []string{}
	publicDNSNames := []string{}
	for _, privateIP := range ni.PrivateIPAddresses {
		privateIPAddresses = append(privateIPAddresses, privateIP.PrivateIPAddress)
		if privateIP.Association.PublicIP != "" {
			publicIPAddresses = append(publicIPAddresses, privateIP.Association.PublicIP)
		}
		if privateIP.Association.PublicDNSName != "" {
			publicDNSNames = append(publicDNSNames, privateIP.Association.PublicDNSName)
		}
	}
	return privateIPAddresses, publicIPAddresses, publicDNSNames
}
