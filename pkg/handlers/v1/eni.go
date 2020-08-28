package v1

import (
	"encoding/json"
	"errors"
	"strings"
)

const elbRequester = "amazon-elb"

type eniConfiguration struct {
	Description        string      `json:"description"`
	PrivateIPAddresses []privateIP `json:"privateIpAddresses"`
	RequesterID        string      `json:"requesterId"`
	RequesterManaged   bool        `json:"requesterManaged"`
}

type eniConfigurationDiff struct {
	PreviousValue *eniConfiguration `json:"previousValue"`
	UpdatedValue  *eniConfiguration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type privateIP struct {
	PrivateIPAddress string `json:"PrivateIpAddress"`
}

type eniTransformer struct{}

func (t eniTransformer) Create(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	var config eniConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, err
	}

	if filter(config) {
		return output, nil
	}

	change := extractEniInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)

	return output, nil
}

// Returns true if we should filter this event due to not being requester managed
func filter(config eniConfiguration) bool {
	return config.RequesterManaged == false || config.RequesterID != elbRequester
}

func (t eniTransformer) Update(event awsConfigEvent) (Output, error) {
	// I don't think requester managed ENIs can update in the traditional sense
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/requester-managed-eni.html
	// This page implies that they can only exist while attached to whatever requested them
	// Because this is effectively a no-op no matter what, we don't need to filter for now

	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	return output, nil
}

func (t eniTransformer) Delete(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	changeProps := event.ConfigurationItemDiff.ChangedProperties

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, errors.New("Invalid configuration diff")
	}
	var configDiff eniConfigurationDiff
	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, err
	}
	if filter(*configDiff.PreviousValue) {
		return output, nil
	}

	change := extractEniInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)
	return output, nil
}

func extractEniInfo(config *eniConfiguration) Change {
	change := Change{}

	// All our sample data only has one privateIp, but it is possible to have multiple
	privateIPs := []string{}
	for _, privateIP := range config.PrivateIPAddresses {
		privateIPs = append(privateIPs, privateIP.PrivateIPAddress)
	}
	change.PrivateIPAddresses = append(change.PrivateIPAddresses, privateIPs...)

	pieces := strings.Split(config.Description, " ")
	change.RelatedResources = append(change.RelatedResources, pieces[len(pieces)-1])

	return change
}
