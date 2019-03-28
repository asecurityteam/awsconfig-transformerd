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
	PreviousValue *[]tag `json:"previousValue"`
	UpdatedValue  *[]tag `json:"updatedValue"`
	ChangeType    string `json:"changeType"`
}

func extractELBNetworkInfo(config *elbConfiguration) Change {
	return Change{
		Hostnames: []string{config.DNSName},
	}
}

func elbOutput(event awsConfigEvent) (Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return Output{}, err
	}
	switch event.ConfigurationItemDiff.ChangeType {
	case create:
		if len(output.Tags) == 0 {
			return Output{}, ErrMissingValue{Field: "Tags"}
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
	case update:
		// DNS names for ELBs cannot be changed, so this case is a largely a no-op.
		if len(output.Tags) == 0 {
			return Output{}, ErrMissingValue{Field: "Tags"}
		}
	case delete:
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
			return Output{}, ErrMissingValue{Field: "ChangedProperties.SupplementaryConfiguration.Tags"}
		}
		var supplementaryConfigDiff supplementaryConfigurationDiff
		if err := json.Unmarshal(supplementaryConfigDiffRaw, &supplementaryConfigDiff); err != nil {
			return Output{}, err
		}
		if supplementaryConfigDiff.PreviousValue == nil || len(*supplementaryConfigDiff.PreviousValue) == 0 {
			return Output{}, ErrMissingValue{Field: "Tags"}
		}
		for _, tag := range *supplementaryConfigDiff.PreviousValue {
			output.Tags[tag.Key] = tag.Value
		}
	default: // NONE
		return Output{}, nil
	}
	return output, nil
}
