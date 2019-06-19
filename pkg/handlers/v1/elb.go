package v1

import (
	"encoding/json"
)

type elbConfiguration struct {
	DNSName string `json:"dnsname"`
}

type elbConfigurationDiff struct {
	PreviousValue *elbConfiguration `json:"previousValue"`
	UpdatedValue  *elbConfiguration `json:"updatedValue"`
	ChangeType    string            `json:"changeType"`
}

type supplementaryConfigurationDiff struct {
	PreviousValue []tag  `json:"previousValue"`
	UpdatedValue  []tag  `json:"updatedValue"`
	ChangeType    string `json:"changeType"`
}

func extractELBNetworkInfo(config *elbConfiguration) Change {
	return Change{
		Hostnames: []string{config.DNSName},
	}
}

type elbTransformer struct{}

func (t elbTransformer) Create(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}

	// if a resource is created for the first time, there is no diff.
	// just read the configuration
	var config elbConfiguration
	if err := json.Unmarshal(event.ConfigurationItem.Configuration, &config); err != nil {
		return Output{}, err
	}
	change := extractELBNetworkInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)
	return output, nil
}

func (t elbTransformer) Update(event awsConfigEvent) (Output, error) {
	// DNS names for ELBs cannot be changed, so the update case is a largely a no-op.
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}
	return output, nil
}

func (t elbTransformer) Delete(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}
	// if a resource is deleted, the tags are no longer present in the base object.
	// we must fetch them from the previous configuration.
	changeProps := event.ConfigurationItemDiff.ChangedProperties
	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return Output{}, ErrMissingValue{Field: "ChangedProperties.Configuration"}
	}
	var configDiff elbConfigurationDiff
	if err := json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return Output{}, err
	}

	// fetch network information from the previous configuration
	change := extractELBNetworkInfo(configDiff.PreviousValue)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)

	// if a resource is deleted, the tags are no longer present in the base object.
	// we must fetch them from the previous configuration.
	supplementaryConfigDiffRaw, ok := changeProps["SupplementaryConfiguration.Tags"]
	if !ok {
		return output, nil
	}
	var supplementaryConfigDiff supplementaryConfigurationDiff
	if err := json.Unmarshal(supplementaryConfigDiffRaw, &supplementaryConfigDiff); err != nil {
		return Output{}, err
	}

	for _, tag := range supplementaryConfigDiff.PreviousValue {
		output.Tags[tag.Key] = tag.Value
	}
	return output, nil
}
