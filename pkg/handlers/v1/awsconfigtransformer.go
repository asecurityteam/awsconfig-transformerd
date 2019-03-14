package v1

import (
	"context" // https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-context.html
	"encoding/json"
	"strings"
	"time"
)

// for "previousValue" and "updatedValue" values under "Configuration.NetworkInterfaces.*"
type ConfigurationNetworkInterfaceValue struct {
	NetworkInterfaceID string `json:"networkInterfaceId"`
	SubnetID           string `json:"subnetId"`
	VpcID              string `json:"vpcId"`
	Description        string `json:"description"`
	OwnerID            string `json:"ownerId"`
	Status             string `json:"status"`
	MacAddress         string `json:"macAddress"`
	PrivateIPAddress   string `json:"privateIpAddress"`
	PrivateDNSName     string `json:"privateDnsName"`
	SourceDestCheck    bool   `json:"sourceDestCheck"`
	Groups             []struct {
		GroupName string `json:"groupName"`
		GroupID   string `json:"groupId"`
	} `json:"groups"`
	Attachment struct {
		AttachmentID        string    `json:"attachmentId"`
		DeviceIndex         int       `json:"deviceIndex"`
		Status              string    `json:"status"`
		AttachTime          time.Time `json:"attachTime"`
		DeleteOnTermination bool      `json:"deleteOnTermination"`
	} `json:"attachment"`
	Association struct {
		PublicIP      string `json:"publicIp"`
		PublicDNSName string `json:"publicDnsName"`
		IPOwnerID     string `json:"ipOwnerId"`
	} `json:"association"`
	PrivateIPAddresses []struct {
		PrivateIPAddress string `json:"privateIpAddress"`
		PrivateDNSName   string `json:"privateDnsName"`
		Primary          bool   `json:"primary"`
		Association      struct {
			PublicIP      string `json:"publicIp"`
			PublicDNSName string `json:"publicDnsName"`
			IPOwnerID     string `json:"ipOwnerId"`
		} `json:"association"`
	} `json:"privateIpAddresses"`
}

// for "previousValue" and "updatedValue" values under "Relationships.*"
type RelationshipValue struct {
	ResourceID   string      `json:"resourceId"`
	ResourceName interface{} `json:"resourceName"` // TODO: string?
	ResourceType string      `json:"resourceType"`
	Name         string      `json:"name"`
}

// for "previousValue" and "updatedValue" values under "Configuration.SecurityGroups..*"
type ConfigurationSecurityGroupValue struct {
	GroupName string `json:"groupName"`
	GroupID   string `json:"groupId"`
}

type ConfigurationNetworkInterface struct {
	PreviousValue *ConfigurationNetworkInterfaceValue `json:"previousValue"`
	UpdatedValue  *ConfigurationNetworkInterfaceValue `json:"updatedValue"`
	ChangeType    string                              `json:"changeType"`
}

type Relationship struct {
	PreviousValue *RelationshipValue `json:"previousValue"`
	UpdatedValue  *RelationshipValue `json:"updatedValue"`
	ChangeType    string             `json:"changeType"`
}

type ConfigurationSecurityGroup struct {
	PreviousValue *ConfigurationSecurityGroupValue `json:"previousValue"`
	UpdatedValue  *ConfigurationSecurityGroupValue `json:"updatedValue"`
	ChangeType    string                           `json:"changeType"`
}

type ConfigurationItemDiff struct {
	ChangedProperties map[string]json.RawMessage `json:"changedProperties` // recommend using getter functions rather than directly accessing
	ChangeType        string                     `json:"changeType"`
}

type AWSConfigEvent struct {
	ConfigurationItemDiff ConfigurationItemDiff `json:"configurationItemDiff"`
	ConfigurationItem     struct {
		RelatedEvents []string            `json:"relatedEvents"`
		Relationships []RelationshipValue `json:"relationships"`
		Configuration struct {
			InstanceID string `json:"instanceId"`
			ImageID    string `json:"imageId"`
			State      struct {
				Code int    `json:"code"`
				Name string `json:"name"`
			} `json:"state"`
			PrivateDNSName        string        `json:"privateDnsName"`
			PublicDNSName         string        `json:"publicDnsName"`
			StateTransitionReason string        `json:"stateTransitionReason"`
			KeyName               string        `json:"keyName"`
			AmiLaunchIndex        int           `json:"amiLaunchIndex"`
			ProductCodes          []interface{} `json:"productCodes"`
			InstanceType          string        `json:"instanceType"`
			LaunchTime            time.Time     `json:"launchTime"`
			Placement             struct {
				AvailabilityZone string      `json:"availabilityZone"`
				GroupName        string      `json:"groupName"`
				Tenancy          string      `json:"tenancy"`
				HostID           interface{} `json:"hostId"`
				Affinity         interface{} `json:"affinity"`
			} `json:"placement"`
			KernelID   interface{} `json:"kernelId"`
			RamdiskID  interface{} `json:"ramdiskId"`
			Platform   interface{} `json:"platform"`
			Monitoring struct {
				State string `json:"state"`
			} `json:"monitoring"`
			SubnetID            string      `json:"subnetId"`
			VpcID               string      `json:"vpcId"`
			PrivateIPAddress    string      `json:"privateIpAddress"`
			PublicIPAddress     string      `json:"publicIpAddress"`
			StateReason         interface{} `json:"stateReason"`
			Architecture        string      `json:"architecture"`
			RootDeviceType      string      `json:"rootDeviceType"`
			RootDeviceName      string      `json:"rootDeviceName"`
			BlockDeviceMappings []struct {
				DeviceName string `json:"deviceName"`
				Ebs        struct {
					VolumeID            string    `json:"volumeId"`
					Status              string    `json:"status"`
					AttachTime          time.Time `json:"attachTime"`
					DeleteOnTermination bool      `json:"deleteOnTermination"`
				} `json:"ebs"`
			} `json:"blockDeviceMappings"`
			VirtualizationType    string      `json:"virtualizationType"`
			InstanceLifecycle     interface{} `json:"instanceLifecycle"`
			SpotInstanceRequestID interface{} `json:"spotInstanceRequestId"`
			ClientToken           string      `json:"clientToken"`
			Tags                  []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"tags"`
			SecurityGroups []struct {
				GroupName string `json:"groupName"`
				GroupID   string `json:"groupId"`
			} `json:"securityGroups"`
			SourceDestCheck    bool                                 `json:"sourceDestCheck"`
			Hypervisor         string                               `json:"hypervisor"`
			NetworkInterfaces  []ConfigurationNetworkInterfaceValue `json:"networkInterfaces"`
			IamInstanceProfile interface{}                          `json:"iamInstanceProfile"`
			EbsOptimized       bool                                 `json:"ebsOptimized"`
			SriovNetSupport    interface{}                          `json:"sriovNetSupport"`
			EnaSupport         bool                                 `json:"enaSupport"`
		} `json:"configuration"`
		SupplementaryConfiguration struct { // TODO: ???
		} `json:"supplementaryConfiguration"`
		Tags struct {
			Name string `json:"Name"`
		} `json:"tags"`
		ConfigurationItemVersion     string      `json:"configurationItemVersion"`
		ConfigurationItemCaptureTime time.Time   `json:"configurationItemCaptureTime"`
		ConfigurationStateID         int64       `json:"configurationStateId"`
		AwsAccountID                 string      `json:"awsAccountId"`
		ConfigurationItemStatus      string      `json:"configurationItemStatus"`
		ResourceType                 string      `json:"resourceType"`
		ResourceID                   string      `json:"resourceId"`
		ResourceName                 interface{} `json:"resourceName"` // TODO: string?
		ARN                          string      `json:"ARN"`
		AwsRegion                    string      `json:"awsRegion"`
		AvailabilityZone             string      `json:"availabilityZone"`
		ConfigurationStateMd5Hash    string      `json:"configurationStateMd5Hash"`
		ResourceCreationTime         time.Time   `json:"resourceCreationTime"`
	} `json:"configurationItem"`
	NotificationCreationTime time.Time `json:"notificationCreationTime"`
	MessageType              string    `json:"messageType"`
	RecordVersion            string    `json:"recordVersion"`
}

func (c *ConfigurationItemDiff) getChangedNetworkInterfaces() map[string]ConfigurationNetworkInterface {

	configurationNetworkInterfaces := make(map[string]ConfigurationNetworkInterface)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Configuration.NetworkInterfaces.") {
			configurationNetworkInterface := ConfigurationNetworkInterface{}
			json.Unmarshal([]byte(value), &configurationNetworkInterface)
			configurationNetworkInterfaces[key] = configurationNetworkInterface
		}
	}
	return configurationNetworkInterfaces
}

func (c *ConfigurationItemDiff) getRelationships() map[string]Relationship {
	relationships := make(map[string]Relationship)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Relationships.") {
			relationship := Relationship{}
			json.Unmarshal([]byte(value), &relationship)
			relationships[key] = relationship
		}
	}
	return relationships
}

func (c *ConfigurationItemDiff) getConfigurationSecurityGroups() map[string]ConfigurationSecurityGroup {
	configurationSecurityGroups := make(map[string]ConfigurationSecurityGroup)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Configuration.SecurityGroups.") {
			configurationSecurityGroup := ConfigurationSecurityGroup{}
			json.Unmarshal([]byte(value), &configurationSecurityGroup)
			configurationSecurityGroups[key] = configurationSecurityGroup
		}
	}
	return configurationSecurityGroups
}

type Output struct {
	PublicIPAddresses  []string `json:"publicIpAddresses"`
	PrivateIPAddresses []string `json:"privateIpAddresses"`
	Hostnames          []string `json:"hostnames"`
	StartedAt          string   `json:"startedAt"` // date-time format
	StoppedAt          string   `json:"stoppedAt"` // date-time format
	ResourceType       string   `json:"resourceType"`
	BusinessUnit       string   `json:"businessUnit"`  // guaranteed
	ResourceOwner      string   `json:"resourceOwner"` // guaranteed
	ServiceName        string   `json:"serviceName"`   // guaranteed
	MicrosServiceID    string   `json:"microsServiceId"`
}

type AWSConfigChangeEventHandler struct {
	Reporter Reporter
}

// function must satisfy AWS lambda.Start parameter spec.  Good enough: https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html
func (h *AWSConfigChangeEventHandler) Handle(ctx context.Context, event AWSConfigEvent) error {
	//return Output{ServiceName: "bob"}, nil // Output{Name: fmt.Sprintf("%s", event.Name)}, nil
	outputs, _ := h.getOutput(&event)
	for _, output := range outputs {
		h.Reporter.Report(ctx, output)
	}
	return nil
}

func (h *AWSConfigChangeEventHandler) getOutput(event *AWSConfigEvent) ([]Output, error) {
	outputs := []Output{}
	outputs = append(outputs, Output{})
	return outputs, nil
}
