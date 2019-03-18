package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Handle is an AWS Lambda handler which takes, as input, an SNS configuration change event notification.
// The input is transformed into a JSON structure which highlights changes in the network details for this resource.
// The output is the transformed JSON.
func Handle(ctx context.Context, input Input) (Output, error) {
	var event awsConfigEvent
	e := json.Unmarshal([]byte(input.Message), &event)
	if e != nil {
		return Output{}, e
	}

	output := Output{
		AccountID:    event.ConfigurationItem.AWSAccountID,
		ChangeTime:   event.ConfigurationItem.ConfigurationItemCaptureTime,
		Region:       event.ConfigurationItem.AWSRegion,
		ResourceID:   event.ConfigurationItem.ResourceID,
		ResourceType: event.ConfigurationItem.ResourceType,
		Tags:         event.ConfigurationItem.Tags,
		Changes:      []Change{}}

	changedNetworkInterfaces := event.ConfigurationItemDiff.getChangedNetworkInterfaces()
	for _, value := range changedNetworkInterfaces {
		configurationNetworkInterfaceValue := value.PreviousValue
		if configurationNetworkInterfaceValue == nil {
			configurationNetworkInterfaceValue = value.UpdatedValue
		}
		change := Change{}
		if value.ChangeType == "DELETE" {
			change.ChangeType = "DELETED"
		} else if value.ChangeType == "CREATE" {
			change.ChangeType = "ADDED"
		}
		change.Hostnames = []string{configurationNetworkInterfaceValue.PrivateDNSName, configurationNetworkInterfaceValue.Association.PublicDNSName}
		change.PrivateIPAddresses = []string{}
		for _, privateIPAddress := range configurationNetworkInterfaceValue.PrivateIPAddresses {
			change.PrivateIPAddresses = append(change.PrivateIPAddresses, privateIPAddress.PrivateIPAddress)
		}
		change.PublicIPAddresses = []string{configurationNetworkInterfaceValue.Association.PublicIP}

		output.Changes = append(output.Changes, change)
	}

	if len(output.Changes) == 0 {
		marshalled, _ := json.Marshal(event)
		return output, errors.New(fmt.Sprintf("Failed to transform AWS Config change event due to lack of sufficient information. The already-marshalled AWS change event was: %s", string(marshalled)))
	}

	return output, nil
}
