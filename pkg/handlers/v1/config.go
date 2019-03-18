package v1

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Input is the event we will receive as input to our lambda handler
type Input struct {
	// Message is the stringified AWS config change notification as documented here:
	// https://docs.aws.amazon.com/config/latest/developerguide/example-sns-notification.html
	Message string `json:"Message"`

	// Timestamp is the time at which the notification was published to the SNS topic
	Timestamp time.Time `json:"Timestamp"`
}

// Output is the result of the transformation
type Output struct {
	// ChangeTime is the time at which the asset change occurred, date-time format (required)
	ChangeTime time.Time `json:"changeTime"`

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

type configurationItemDiff struct {
	ChangedProperties map[string]json.RawMessage `json:"changedProperties"` // recommend using getter functions rather than directly accessing
	ChangeType        string                     `json:"changeType"`
}

type awsConfigEvent struct {
	ConfigurationItemDiff configurationItemDiff `json:"configurationItemDiff"`
	ConfigurationItem     struct {
		Configuration                *json.RawMessage  `json:"configuration"`
		RelatedEvents                []string          `json:"relatedEvents"`
		Relationships                []relationship    `json:"relationships"`
		SupplementaryConfiguration   map[string]string `json:"supplementaryConfiguration"`
		Tags                         map[string]string `json:"tags"`
		ConfigurationItemVersion     string            `json:"configurationItemVersion"`
		ConfigurationItemCaptureTime time.Time         `json:"configurationItemCaptureTime"`
		ConfigurationStateID         int64             `json:"configurationStateId"`
		AWSAccountID                 string            `json:"awsAccountId"`
		ConfigurationItemStatus      string            `json:"configurationItemStatus"`
		ResourceType                 string            `json:"resourceType"`
		ResourceID                   string            `json:"resourceId"`
		ResourceName                 interface{}       `json:"resourceName"`
		ARN                          string            `json:"ARN"`
		AWSRegion                    string            `json:"awsRegion"`
		AvailabilityZone             string            `json:"availabilityZone"`
		ConfigurationStateMd5Hash    string            `json:"configurationStateMd5Hash"`
		ResourceCreationTime         time.Time         `json:"resourceCreationTime"`
	} `json:"configurationItem"`
	NotificationCreationTime time.Time `json:"notificationCreationTime"`
	MessageType              string    `json:"messageType"`
	RecordVersion            string    `json:"recordVersion"`
}

type relationship struct {
	ResourceID   string      `json:"resourceId"`
	ResourceName interface{} `json:"resourceName"`
	ResourceType string      `json:"resourceType"`
	Name         string      `json:"name"`
}

func (c *configurationItemDiff) getChangedNetworkInterfaces() []configurationNetworkInterface {

	configurationNetworkInterfaces := make(map[int]*configurationNetworkInterface)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Configuration.NetworkInterfaces.") {
			// we parse the index because JSON key order is not guaranteed,
			// and by using the index from the original key, we preserve array
			// order and make testing simpler and deterministic
			stringSlice := strings.Split(key, ".")
			index, _ := strconv.Atoi(stringSlice[len(stringSlice)-1])
			configurationNetworkInterface := configurationNetworkInterface{}
			json.Unmarshal([]byte(value), &configurationNetworkInterface)
			configurationNetworkInterfaces[index] = &configurationNetworkInterface
		}
	}

	// make an ordered array
	configurationNetworkInterfacesArray := []configurationNetworkInterface{}
	i := 0
	for {
		c := configurationNetworkInterfaces[i]
		if c == nil {
			break
		}
		configurationNetworkInterfacesArray = append(configurationNetworkInterfacesArray, *c)
		i++
	}
	return configurationNetworkInterfacesArray
}
