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
	Tags                  []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"tags"`
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
}

type networkInterfaceDiff struct {
	PreviousValue *networkInterface `json:"previousValue"`
	UpdatedValue  *networkInterface `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type configurationDiff struct {
	PreviousValue *ec2Configuration `json:"previousValue"`
	UpdatedValue  *ec2Configuration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

func ec2Output(event awsConfigEvent) (Output, error) {
	output := getBaseOutput(event.ConfigurationItem)
	switch event.ConfigurationItemDiff.ChangeType {
	case create:
		// if a resource is created for the first time, there is no diff.
		// just read the configuration
		var config ec2Configuration
		if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
			return Output{}, err
		}
		change := extractEC2NetworkInfo(&config)
		change.ChangeType = added
		output.Changes = append(output.Changes, change)
	case update:
		addedChanges := Change{ChangeType: added}
		removedChanges := Change{ChangeType: deleted}
		// If an update was detected, check to see if any changes to the NetworkInterfaces occurred
		for k, v := range event.ConfigurationItemDiff.ChangedProperties {
			if !strings.HasPrefix(k, "Configuration.NetworkInterfaces.") {
				continue
			}
			var diff networkInterfaceDiff
			if err := json.Unmarshal(v, &diff); err != nil {
				return Output{}, err
			}
			ni := diff.UpdatedValue
			changes := &addedChanges
			if diff.ChangeType == delete {
				ni = diff.PreviousValue
				changes = &removedChanges
			}
			private, public, dns := extractNetworkInterfaceInfo(ni)
			changes.PrivateIPAddresses = append(changes.PrivateIPAddresses, private...)
			changes.PublicIPAddresses = append(changes.PublicIPAddresses, public...)
			changes.Hostnames = append(changes.Hostnames, dns...)
		}

		// We need to dedup across the sets of changes to reduce redundancy.
		// i.e. remove entries that show up as both added and removed
		removeDuplicates(&addedChanges, &removedChanges)
		if len(addedChanges.PrivateIPAddresses) > 0 || len(addedChanges.PublicIPAddresses) > 0 || len(addedChanges.Hostnames) > 0 {
			output.Changes = append(output.Changes, addedChanges)
		}
		if len(removedChanges.PrivateIPAddresses) > 0 || len(removedChanges.PublicIPAddresses) > 0 || len(removedChanges.Hostnames) > 0 {
			output.Changes = append(output.Changes, removedChanges)
		}
	case delete:
		// if a resource is deleted, the tags are no longer present in the base object.
		// we must fetch them from the previous configuration.
		changeProps := event.ConfigurationItemDiff.ChangedProperties
		configDiffRaw, ok := changeProps["Configuration"]
		if !ok {
			return Output{}, errors.New("Invalid configuration diff")
		}
		var configDiff configurationDiff
		if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
			return Output{}, err
		}
		for _, tag := range configDiff.PreviousValue.Tags {
			output.Tags[tag.Key] = tag.Value
		}

		// fetch network information from the previous configuration
		change := extractEC2NetworkInfo(configDiff.PreviousValue)
		change.ChangeType = deleted
		output.Changes = append(output.Changes, change)
	default: // NONE
		return Output{}, nil
	}
	return output, nil
}

func removeDuplicates(a, b *Change) {
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
	for _, ni := range config.NetworkInterfaces {
		private, public, dns := extractNetworkInterfaceInfo(&ni)
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
