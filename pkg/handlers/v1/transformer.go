package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/configservice"

	"github.com/asecurityteam/awsconfig-transformerd/pkg/domain"
	"github.com/asecurityteam/awsconfig-transformerd/pkg/logs"
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

	// ProcessedTimestamp is an optional field. It is the time at which a previous service emitted this event.
	ProcessedTimestamp string `json:"ProcessedTimestamp"`
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

	// ARN is the Amazon Resource Name (required)
	ARN string `json:"arn"`

	// Tags are key/value pairs set on the AWS resource (required)
	Tags map[string]string `json:"tags"`

	// Changes are a list of network related changes which occurred on the resource (required)
	Changes []Change `json:"changes"`
}

// Change details network related changes for a resource
type Change struct {
	// PublicIPAddresses show changed public IP addresses
	PublicIPAddresses []string `json:"publicIpAddresses,omitempty"`

	// PrivateIPAddresses show changed private IP addresses
	PrivateIPAddresses []string `json:"privateIpAddresses,omitempty"`

	// Hostnames show changed public DNS names
	Hostnames []string `json:"hostnames,omitempty"`

	// RelatedResources show a related arn_id. ex: an ELB the ENI is attached to
	RelatedResources []string `json:"relatedResources,omitempty"`

	// TagChanges changed keys/values per tag
	TagChanges []TagChange `json:"tagChanges,omitempty"`

	// ChangeType indicates the type of change which occurred. Allowed values are "ADDED" or "DELETED"
	ChangeType string `json:"changeType"`
}

// TagChange represents a modification, addition or deletion of a resource tag key or value
type TagChange struct {
	UpdatedValue  *Tag `json:"updatedValue"` // pointer type as either of the values can be nil
	PreviousValue *Tag `json:"previousValue"`
}

// Tag represents a single AWS resource tag (key:value pair)
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Transformer is a lambda handler which transforms incoming AWS Config change events
type Transformer struct {
	LogFn  domain.LogFn
	StatFn domain.StatFn
}

// Handle is an AWS Lambda handler which takes, as input, an SNS configuration change event notification.
// The input is transformed into a JSON structure which highlights changes in the network details for this resource.
// The output is the transformed JSON.
func (t *Transformer) Handle(ctx context.Context, input Input) ([]Output, error) {

	if ts, err := time.Parse(time.RFC3339Nano, input.ProcessedTimestamp); err == nil {
		t.StatFn(ctx).Timing("event.awsconfig.transformer.event.delay", time.Since(ts))
	}

	var event awsConfigEvent
	err := json.Unmarshal([]byte(input.Message), &event)
	if err != nil {
		t.LogFn(ctx).Error(logs.TransformError{Reason: err.Error()})
		return []Output{}, err
	}

	var outputs []Output

	switch event.ConfigurationItem.ResourceType {
	case configservice.ResourceTypeAwsEc2Instance:
		outputs, err = transformOutput(event, ec2Transformer{})
	case configservice.ResourceTypeAwsElasticLoadBalancingLoadBalancer:
		outputs, err = transformOutput(event, elbTransformer{})
	case configservice.ResourceTypeAwsElasticLoadBalancingV2LoadBalancer:
		// ALB Config events have the same as ELBs
		outputs, err = transformOutput(event, elbTransformer{})
	case configservice.ResourceTypeAwsEc2NetworkInterface:
		outputs, err = transformOutput(event, eniTransformer{})
	default:
		t.LogFn(ctx).Info(logs.UnsupportedResource{Resource: event.ConfigurationItem.ResourceType})
		err = errors.New("Unsupported resource type; must be one of EC2, ELB, ENI")
	}

	if err != nil {
		t.LogFn(ctx).Error(logs.TransformError{Reason: err.Error()})
		// do not proceed to extract tags if the event is broken/malformed
		return []Output{}, err
	}

	tagChanges, err := extractTagChanges(event.ConfigurationItemDiff)
	if err != nil {
		t.LogFn(ctx).Error(logs.TransformError{Reason: err.Error()})
		return []Output{}, err
	}
	if len(tagChanges) > 0 {
		op := added
		if event.MessageType == delete {
			op = deleted
		}
		for _, output := range outputs {
			output.Changes = append(output.Changes, Change{
				TagChanges: tagChanges,
				ChangeType: op,
			})
		}
	}
	return outputs, err
}

func extractTagChanges(ev configurationItemDiff) ([]TagChange, error) {
	res := make([]TagChange, 0)
	for k, v := range ev.ChangedProperties {
		if !strings.HasPrefix(k, "Configuration.TagSet.") &&
			!strings.HasPrefix(k, "SupplementaryConfiguration.TagSet.") &&
			!strings.HasPrefix(k, "TagSet.") {
			continue
		}
		var tc TagChange
		if err := json.Unmarshal(v, &tc); err != nil {
			return nil, err
		}
		if tc.PreviousValue == nil && tc.UpdatedValue == nil {
			return nil, fmt.Errorf("malformed tag change event")
		}
		res = append(res, tc)
	}
	return res, nil
}

// ResourceTransformer takes AWS Config Events, and returns transformed
// output.
type ResourceTransformer interface {
	Create(event awsConfigEvent) ([]Output, error)
	Update(event awsConfigEvent) ([]Output, error)
	Delete(event awsConfigEvent) ([]Output, error)
}

func transformOutput(event awsConfigEvent, resourceTransformer ResourceTransformer) ([]Output, error) {
	switch event.ConfigurationItemDiff.ChangeType {
	case create:
		output, err := resourceTransformer.Create(event)
		if err != nil {
			return []Output{}, err
		}
		return output, nil
	case update:
		output, err := resourceTransformer.Update(event)
		if err != nil {
			return []Output{}, err
		}
		return output, nil
	case delete:
		output, err := resourceTransformer.Delete(event)
		if err != nil {
			return []Output{}, err
		}
		return output, nil
	}
	return []Output{}, errors.New("Event was not create, update, or delete")
}
