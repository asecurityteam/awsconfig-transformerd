package v1

import (
	"encoding/json"
	"time"
)

type elbConfiguration struct {
	DNSName     string `json:"dnsname"`
	CreatedTime string `json:"createdTime"`
}

type elbConfigurationMillis struct {
	DNSName     string `json:"dnsname"`
	CreatedTime int64  `json:"createdTime"`
}

type elbConfigurationDiff struct {
	PreviousValue json.RawMessage `json:"previousValue"` // elbConfiguration
	UpdatedValue  json.RawMessage `json:"updatedValue"`  // elbConfiguration
	ChangeType    string          `json:"changeType"`
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

// Some ELBs have createdTime in the form of milliseconds, which is int64 in Go,
// so we need to convert those to a RFC3339 timestamp string. This unmarshal function
// returns an elbConfiguration that uses RFC3339 timestamp string with milliseconds, no matter
// what kind of timestamp the original config had.
func unmarshalConfig(rawConfig json.RawMessage) (elbConfiguration, error) {
	var config elbConfiguration
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		// Timestamp came in form of milliseconds, we must convert to timestamp string
		var configMillis elbConfigurationMillis
		if err = json.Unmarshal(rawConfig, &configMillis); err != nil {
			return elbConfiguration{}, err
		}
		config.DNSName = configMillis.DNSName
		config.CreatedTime = time.Unix(0, configMillis.CreatedTime*int64(time.Millisecond)).Format("2006-01-02T15:04:05.000Z")
	}
	return config, nil
}

func (t elbTransformer) Create(event awsConfigEvent) ([]Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}
	// if a resource is created for the first time, there is no diff.
	// just read the configuration
	config, err := unmarshalConfig(event.ConfigurationItem.Configuration)
	if err != nil {
		return []Output{}, err
	}

	change := extractELBNetworkInfo(&config)
	change.ChangeType = added
	output.Changes = append(output.Changes, change)
	output.ChangeTime = config.CreatedTime
	return []Output{output}, nil
}

func (t elbTransformer) Update(event awsConfigEvent) ([]Output, error) {
	// DNS names for ELBs cannot be changed, so the update case is a largely a no-op.
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}
	config, err := unmarshalConfig(event.ConfigurationItem.Configuration)
	if err != nil {
		return []Output{}, err
	}
	output.ChangeTime = config.CreatedTime
	return []Output{output}, nil
}

func (t elbTransformer) Delete(event awsConfigEvent) ([]Output, error) {
	output, err := getBaseOutput(event.ConfigurationItem)
	if err != nil {
		return []Output{}, err
	}
	// if a resource is deleted, the tags are no longer present in the base object.
	// we must fetch them from the previous configuration.
	changeProps := event.ConfigurationItemDiff.ChangedProperties
	configDiffRaw, ok := changeProps["Configuration"]
	if !ok {
		return []Output{}, ErrMissingValue{Field: "ChangedProperties.Configuration"}
	}
	var configDiff elbConfigurationDiff
	if err = json.Unmarshal(configDiffRaw, &configDiff); err != nil {
		return []Output{}, err
	}

	config, err := unmarshalConfig(configDiff.PreviousValue)
	if err != nil {
		return []Output{}, err
	}

	// fetch network information from the previous configuration
	change := extractELBNetworkInfo(&config)
	change.ChangeType = deleted
	output.Changes = append(output.Changes, change)

	// if a resource is deleted, the tags are no longer present in the base object.
	// we must fetch them from the previous configuration.
	supplementaryConfigDiffRaw, ok := changeProps["SupplementaryConfiguration.Tags"]
	if !ok {
		return []Output{output}, nil
	}
	var supplementaryConfigDiff supplementaryConfigurationDiff
	if err := json.Unmarshal(supplementaryConfigDiffRaw, &supplementaryConfigDiff); err != nil {
		return []Output{}, err
	}

	for _, tag := range supplementaryConfigDiff.PreviousValue {
		output.Tags[tag.Key] = tag.Value
	}
	return []Output{output}, nil
}
