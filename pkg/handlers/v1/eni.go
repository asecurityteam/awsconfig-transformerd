package v1

import (
	"encoding/json"
	"errors"
	"strings"
)

type eniConfiguration struct {
	Description			string		`json:"description"`
	PrivateIPAddresses 	[]privateIp	`json:"privateIpAddresses"`
	RequesterId			string		`json:"requesterId"`
	RequesterManaged	string		`json:"requesterManaged"`
}

type eniConfigurationDiff struct {
	PreviousValue *eniConfiguration `json:"previousValue"`
	UpdatedValue  *eniConfiguration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type privateIp struct {
	PrivateIpAddress	string		`json:"PrivateIpAddress"`
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
	change := extractEniInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)

	// TODO: Check requesterId and requesterManaged. If they fail can I return Output{} and be fine?
	// Check with Denise maybe about how EC2 events are filtered? I couldn't figure it out at a glance

	return output, nil
}

func (t eniTransformer) Update(event awsConfigEvent) (Output, error) {
	// I don't think requesterManaged ENIs can update in the traditional sense
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/requester-managed-eni.html
	// This page implies that they can only exist while attached to whatever requested them

	// Does this have unintended consequences?
	return Output{}, nil
}

func (t eniTransformer) Delete(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	changeProps := event.ConfigurationItemDiff.ChangedProperties
	// TODO: Can ARN be null in the ENI config like EC2?

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, errors.New("Invalid configuration diff")
	}
	var configDiff eniConfigurationDiff
	if err:= json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, err
	}

	change := extractEniInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)
	return output, nil
}

func extractEniInfo(config *eniConfiguration) Change {
	change := Change{}

	// All our sample data only has one privateIp, but it is possible to have multiple
	privateIps := []string{}
	for _, privateIp := range config.PrivateIPAddresses {
		privateIps = append(privateIps, privateIp.PrivateIpAddress)
	}
	change.PrivateIPAddresses = append(change.PrivateIPAddresses, privateIps...)

	// I in general am not a fan of this, but the only other thing I could think of was
	// regex matching which doesn't seem any better
	pieces := strings.Split(config.Description, " ")
	change.RelatedResource = append(change.RelatedResource, pieces[len(pieces) -1])

	return change
}