package v1

import (
	"encoding/json"
	"errors"
	"strings"
)

const elbRequester = "amazon-elb"

type eniConfiguration struct {
	Description        string             `json:"description"`
	PrivateIPAddresses []privateIPAddress `json:"privateIpAddresses"`
	RequesterID        string             `json:"requesterId"`
	RequesterManaged   bool               `json:"requesterManaged"`
}

type eniConfigurationDiff struct {
	PreviousValue *eniConfiguration `json:"previousValue"`
	UpdatedValue  *eniConfiguration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type privateIPBlockDiff struct {
	PreviousValue *privateIPAddress `json:"previousValue"`
	UpdatedValue  *privateIPAddress `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type privateIPAddress struct {
	PrivateIPAddress string `json:"privateIpAddress"`
	PrivateDNSName   string `json:"privateDnsName"`
	Primary          bool   `json:"primary"`
	Association      struct {
		PublicIP      string `json:"publicIp"`
		PublicDNSName string `json:"publicDnsName"`
		IPOwnerID     string `json:"ipOwnerId"`
	} `json:"association"`
}

type eniTransformer struct{}

func (t eniTransformer) Create(event awsConfigEvent) (Output, bool, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	var config eniConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, false, err
	}

	if filter(config) {
		return output, true, nil
	}

	change := extractEniInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)

	return output, false, nil
}

// Returns true if we should filter this event due to not being requester managed
func filter(config eniConfiguration) bool {
	return !config.RequesterManaged || config.RequesterID != elbRequester
}

func (t eniTransformer) Update(event awsConfigEvent) (Output, bool, error) {
	// I don't think requester managed ENIs can update in the traditional sense
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/requester-managed-eni.html
	// This page implies that they can only exist while attached to whatever requested them
	// Because this is effectively a no-op no matter what, we don't need to filter for now
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	var config eniConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, false, err
	}

	//return output, filter(config), nil

	addedChange := Change{ChangeType: added}
	deletedChange := Change{ChangeType: deleted}
	// If an update was detected, check to see if any changes to the NetworkInterfaces occurred
	for k, v := range event.ConfigurationItemDiff.ChangedProperties {
		if !strings.HasPrefix(k, "Configuration.PrivateIpAddresses.") {
			continue
		}
		var diff privateIPBlockDiff
		if err := json.Unmarshal(v, &diff); err != nil {
			return Output{}, false, err
		}
		ipBlock := diff.UpdatedValue
		changes := &addedChange
		if diff.ChangeType == delete {
			ipBlock = diff.PreviousValue
			changes = &deletedChange
		}
		extractIPBlock(ipBlock, changes)
		extractRelatedResources(&config, changes)
	}
	// We need to compute the symmetric difference of the added changes and the removed changes
	// i.e. remove entries that show up as both added and removed
	symmetricDifference(&addedChange, &deletedChange)
	if len(addedChange.PrivateIPAddresses) > 0 || len(addedChange.PublicIPAddresses) > 0 || len(addedChange.Hostnames) > 0 {
		output.Changes = append(output.Changes, addedChange)
	}
	if len(deletedChange.PrivateIPAddresses) > 0 || len(deletedChange.PublicIPAddresses) > 0 || len(deletedChange.Hostnames) > 0 {
		output.Changes = append(output.Changes, deletedChange)
	}
	return output, false, nil
}

func (t eniTransformer) Delete(event awsConfigEvent) (Output, bool, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	changeProps := event.ConfigurationItemDiff.ChangedProperties

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, false, errors.New("Invalid configuration diff")
	}
	var configDiff eniConfigurationDiff
	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, false, err
	}
	if filter(*configDiff.PreviousValue) {
		return output, true, nil
	}

	change := extractEniInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)
	return output, false, nil
}

func extractEniInfo(config *eniConfiguration) Change {
	change := Change{}

	for _, privateIP := range config.PrivateIPAddresses {
		extractIPBlock(&privateIP, &change)
	}

	extractRelatedResources(config, &change)

	return change
}

func extractRelatedResources(config *eniConfiguration, change *Change) {
	pieces := strings.Split(config.Description, " ")
	change.RelatedResources = append(change.RelatedResources, pieces[len(pieces)-1])
}

func extractIPBlock(privateIpBlock *privateIPAddress, change *Change) {
	change.PrivateIPAddresses = append(change.PrivateIPAddresses, privateIpBlock.PrivateIPAddress)
	if privateIpBlock.Association.PublicIP != "" {
		change.PublicIPAddresses = append(change.PublicIPAddresses, privateIpBlock.Association.PublicIP)
	}
	if privateIpBlock.Association.PublicDNSName != "" {
		change.Hostnames = append(change.Hostnames, privateIpBlock.Association.PublicDNSName)
	}
}
