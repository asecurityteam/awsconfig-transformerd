package v1

import (
	"encoding/json"
	"errors"
	"fmt"
)

type subnetConfiguration struct {
	CIDRBlock string `json:"cidrBlock"`
	VPCID     string `json:"vpcId"`
}

type subnetConfigurationDiff struct {
	PreviousValue *subnetConfiguration `json:"previousValue"`
	UpdatedValue  *subnetConfiguration `json:"updatedValue"`
	ChangeType    string               `json:"changeType"`
}

type subnetTransformer struct{}

func (t subnetTransformer) Create(event awsConfigEvent) (Output, bool, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	var config subnetConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, false, err
	}

	change := extractSubnetInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)
	return output, false, nil
}

func (t subnetTransformer) Update(event awsConfigEvent) (Output, bool, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	return output, false, nil
}

func (t subnetTransformer) Delete(event awsConfigEvent) (Output, bool, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, false, err
	}

	changeProps := event.ConfigurationItemDiff.ChangedProperties

	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, false, errors.New("invalid configuration diff")
	}

	var configDiff subnetConfigurationDiff

	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, false, err
	}

	change := extractSubnetInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)
	fmt.Println(output)
	return output, false, nil
}

func extractSubnetInfo(config *subnetConfiguration) Change {
	change := Change{}
	change.CIDRBlock = config.CIDRBlock
	change.RelatedResources = append(change.RelatedResources, config.VPCID)
	return change
}
