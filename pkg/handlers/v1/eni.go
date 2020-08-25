package v1

import (
	"encoding/json"
	"errors"
	"strings"
)

const elbManaged = "amazon-elb"

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

	if config.RequesterManaged == false || config.RequesterID != elbManaged {
		return Output{}, nil
	}
	change := extractEniInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)

	return output, nil
}

func (t eniTransformer) Update(event awsConfigEvent) (Output, error) {
	// I don't think requesterManaged ENIs can update in the traditional sense
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/requester-managed-eni.html
	// This page implies that they can only exist while attached to whatever requested them

	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	// We don't care about this config outside of checking if it's requesterManaged or has the right requester type
	var config eniConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, err
	}
	if config.RequesterManaged == false || config.RequesterID != elbManaged {
		return Output{}, nil
	}
	return output, nil
}

func (t eniTransformer) Delete(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	changeProps := event.ConfigurationItemDiff.ChangedProperties
	// TODO: Can ARN be null in the ENI config like EC2?
	// TODO: Get a sample requesterManaged ENI event with tags. I'm not sure if we need tagSet for previousValue

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, errors.New("Invalid configuration diff")
	}
	var configDiff eniConfigurationDiff
	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, err
	}

	if configDiff.PreviousValue.RequesterManaged == false || configDiff.PreviousValue.RequesterID != elbManaged {
		return Output{}, nil
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

	// I in general am not a fan of this, but the only other thing I could think of was
	// regex matching which doesn't seem any better
	pieces := strings.Split(config.Description, " ")
	change.RelatedResources = append(change.RelatedResources, pieces[len(pieces)-1])

	return change
}
