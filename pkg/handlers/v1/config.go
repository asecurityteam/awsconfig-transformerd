package v1

import (
	"encoding/json"
)

const (
	// AWS Config change types
	create = "CREATE"
	delete = "DELETE"
	update = "UPDATE"
)

type configurationItemDiff struct {
	ChangedProperties map[string]json.RawMessage `json:"changedProperties"`
	ChangeType        string                     `json:"changeType"`
}

type configurationItem struct {
	Configuration                json.RawMessage        `json:"configuration"`
	RelatedEvents                []string               `json:"relatedEvents"`
	Relationships                []relationship         `json:"relationships"`
	SupplementaryConfiguration   map[string]interface{} `json:"supplementaryConfiguration"`
	Tags                         map[string]string      `json:"tags"`
	ConfigurationItemVersion     string                 `json:"configurationItemVersion"`
	ConfigurationItemCaptureTime string                 `json:"configurationItemCaptureTime"`
	ConfigurationStateID         int64                  `json:"configurationStateId"`
	AWSAccountID                 string                 `json:"awsAccountId"`
	ConfigurationItemStatus      string                 `json:"configurationItemStatus"`
	ResourceType                 string                 `json:"resourceType"`
	ResourceID                   string                 `json:"resourceId"`
	ResourceName                 interface{}            `json:"resourceName"`
	ARN                          string                 `json:"ARN"`
	AWSRegion                    string                 `json:"awsRegion"`
	AvailabilityZone             string                 `json:"availabilityZone"`
	ConfigurationStateMd5Hash    string                 `json:"configurationStateMd5Hash"`
	ResourceCreationTime         string                 `json:"resourceCreationTime"`
}

type awsConfigEvent struct {
	ConfigurationItemDiff    configurationItemDiff `json:"configurationItemDiff"`
	ConfigurationItem        configurationItem     `json:"configurationItem"`
	NotificationCreationTime string                `json:"notificationCreationTime"`
	MessageType              string                `json:"messageType"`
	RecordVersion            string                `json:"recordVersion"`
}

type relationship struct {
	ResourceID   string      `json:"resourceId"`
	ResourceName interface{} `json:"resourceName"`
	ResourceType string      `json:"resourceType"`
	Name         string      `json:"name"`
}

func getBaseOutput(c configurationItem) (Output, error) {
	if c.AWSAccountID == "" {
		return Output{}, ErrMissingValue{Field: "AWSAccountID"}
	}
	if c.AWSRegion == "" {
		return Output{}, ErrMissingValue{Field: "AWSRegion"}
	}
	if c.ConfigurationItemCaptureTime == "" {
		return Output{}, ErrMissingValue{Field: "ConfigurationItemCaptureTime"}
	}
	if c.ResourceType == "" {
		return Output{}, ErrMissingValue{Field: "ResourceType"}
	}
	return Output{
		AccountID:    c.AWSAccountID,
		ChangeTime:   c.ConfigurationItemCaptureTime,
		Region:       c.AWSRegion,
		ResourceType: c.ResourceType,
		ARN:          c.ARN,
		Tags:         c.Tags,
	}, nil
}
