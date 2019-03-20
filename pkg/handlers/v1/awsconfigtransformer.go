package v1

import (
	"context"
	"encoding/json"

	"bitbucket.org/asecurityteam/awsconfig-transformerd/pkg/domain"
	"bitbucket.org/asecurityteam/awsconfig-transformerd/pkg/logs"
	"github.com/aws/aws-sdk-go/service/configservice"
)

const (
	// change types for converted output
	added   = "ADDED"
	deleted = "DELETED"
)

// Input is the event we will receive as input to our lambda handler
type Input struct {
	// Message is the stringified AWS config change notification as documented here:
	// https://docs.aws.amazon.com/config/latest/developerguide/example-sns-notification.html
	Message string `json:"Message"`

	// Timestamp is the time at which the notification was published to the SNS topic
	Timestamp string `json:"Timestamp"`
}

// Output is the result of the transformation
type Output struct {
	// ChangeTime is the time at which the asset change occurred, date-time format (required)
	ChangeTime string `json:"changeTime"`

	// ResourceType is the AWS resource type (required)
	ResourceType string `json:"resourceType"`

	// AccountID is the 12-digit ID of the AWS account (required)
	AccountID string `json:"accountId"`

	// Region is the AWS region (required)
	Region string `json:"region"`

	// ResourceID is the ID of the AWS resource (required)
	ResourceID string `json:"resourceId"`

	// Tags are key/value pairs set on the AWS resource (required)
	Tags map[string]string `json:"tags"`

	// Changes are a list of network related changes which occurred on the resource (required)
	Changes []Change `json:"changes"`
}

// Change details network related changes for a resource
type Change struct {
	// PublicIPAddresses show changed public IP addresses
	PublicIPAddresses []string `json:"publicIpAddresses"`

	// PrivateIPAddresses show changed private IP addresses
	PrivateIPAddresses []string `json:"privateIpAddresses"`

	// Hostnames show changed public DNS names
	Hostnames []string `json:"hostnames"`

	// ChangeType indicates the type of change which occurred. Allowed values are "ADDED" or "DELETED"
	ChangeType string `json:"changeType"`
}

// Transformer is a lambda handler which transforms incoming AWS Config change events
type Transformer struct {
	LogFn  domain.LogFn
	StatFn domain.StatFn
}

// Handle is an AWS Lambda handler which takes, as input, an SNS configuration change event notification.
// The input is transformed into a JSON structure which highlights changes in the network details for this resource.
// The output is the transformed JSON.
func (t *Transformer) Handle(ctx context.Context, input Input) (Output, error) {
	var event awsConfigEvent
	e := json.Unmarshal([]byte(input.Message), &event)
	if e != nil {
		return Output{}, e
	}

	switch event.ConfigurationItem.ResourceType {
	case configservice.ResourceTypeAwsEc2Instance:
		return ec2Output(event)
	default:
		t.LogFn(ctx).Info(logs.UnsupportedResource{Resource: event.ConfigurationItem.ResourceType})
	}

	// if there is nothing to transform, return an empty output
	return Output{}, nil
}
