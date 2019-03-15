package v1

import (
	"context" // https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-context.html
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
		Tags                         map[string]string `json:"tags"`
		ConfigurationItemVersion     string            `json:"configurationItemVersion"`
		ConfigurationItemCaptureTime time.Time         `json:"configurationItemCaptureTime"`
		ConfigurationStateID         int64             `json:"configurationStateId"`
		AwsAccountID                 string            `json:"awsAccountId"`
		ConfigurationItemStatus      string            `json:"configurationItemStatus"`
		ResourceType                 string            `json:"resourceType"`
		ResourceID                   string            `json:"resourceId"`
		ResourceName                 interface{}       `json:"resourceName"` // TODO: string?
		ARN                          string            `json:"ARN"`
		AwsRegion                    string            `json:"awsRegion"`
		AvailabilityZone             string            `json:"availabilityZone"`
		ConfigurationStateMd5Hash    string            `json:"configurationStateMd5Hash"`
		ResourceCreationTime         time.Time         `json:"resourceCreationTime"`
	} `json:"configurationItem"`
	NotificationCreationTime time.Time `json:"notificationCreationTime"`
	MessageType              string    `json:"messageType"`
	RecordVersion            string    `json:"recordVersion"`
}

func (c *ConfigurationItemDiff) getChangedNetworkInterfaces() []ConfigurationNetworkInterface {

	configurationNetworkInterfaces := make(map[int]*ConfigurationNetworkInterface)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Configuration.NetworkInterfaces.") {
			// we parse the index because JSON key order is not guaranteed,
			// and by using the index from the original key, we preserve array
			// order and make testing simpler and deterministic
			stringSlice := strings.Split(key, ".")
			index, _ := strconv.Atoi(stringSlice[len(stringSlice)-1])
			configurationNetworkInterface := ConfigurationNetworkInterface{}
			json.Unmarshal([]byte(value), &configurationNetworkInterface)
			configurationNetworkInterfaces[index] = &configurationNetworkInterface
		}
	}

	// make an ordered array
	configurationNetworkInterfacesArray := []ConfigurationNetworkInterface{}
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

func (c *ConfigurationItemDiff) getRelationships() []Relationship {
	relationships := make(map[int]*Relationship)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Relationships.") {
			// we parse the index because JSON key order is not guaranteed,
			// and by using the index from the original key, we preserve array
			// order and make testing simpler and deterministic
			stringSlice := strings.Split(key, ".")
			index, _ := strconv.Atoi(stringSlice[len(stringSlice)-1])
			relationship := Relationship{}
			json.Unmarshal([]byte(value), &relationship)
			relationships[index] = &relationship
		}
	}

	// make an ordered array
	relationshipsArray := []Relationship{}
	i := 0
	for {
		r := relationships[i]
		if r == nil {
			break
		}
		relationshipsArray = append(relationshipsArray, *r)
		i++
	}
	return relationshipsArray
}

func (c *ConfigurationItemDiff) getConfigurationSecurityGroups() []ConfigurationSecurityGroup {
	configurationSecurityGroups := make(map[int]*ConfigurationSecurityGroup)
	for key, value := range c.ChangedProperties {
		if strings.HasPrefix(key, "Configuration.SecurityGroups.") {
			// we parse the index because JSON key order is not guaranteed,
			// and by using the index from the original key, we preserve array
			// order and make testing simpler and deterministic
			stringSlice := strings.Split(key, ".")
			index, _ := strconv.Atoi(stringSlice[len(stringSlice)-1])
			configurationSecurityGroup := ConfigurationSecurityGroup{}
			json.Unmarshal([]byte(value), &configurationSecurityGroup)
			configurationSecurityGroups[index] = &configurationSecurityGroup
		}
	}

	// make an ordered array
	configurationSecurityGroupsArray := []ConfigurationSecurityGroup{}
	i := 0
	for {
		r := configurationSecurityGroups[i]
		if r == nil {
			break
		}
		configurationSecurityGroupsArray = append(configurationSecurityGroupsArray, *r)
		i++
	}
	return configurationSecurityGroupsArray
}

type Change struct {
	PublicIPAddresses  []string `json:"publicIpAddresses"`
	PrivateIPAddresses []string `json:"privateIpAddresses"`
	Hostnames          []string `json:"hostnames"`
	ChangeType         string   `json:"changeType"` // values are "ADDED" or "DELETED"
}

// TODO: document which parts are required, per https://bitbucket.org/asecurityteam/secdev-docs/src/64ec70b2e544eb889ac3907a5715e97777e35e34/content/platform/security/refs/arch/assetinventory.md?at=SECD-214
type Output struct {
	ChangeTime   time.Time         `json:"changeTime"`   // time at which the asset change occurred, date-time format
	ResourceType string            `json:"resourceType"` // the AWS resource type
	AccountID    string            `json:"accountId"`    // the ID of the AWS account
	Region       string            `json:"region"`       // the AWS region
	ResourceID   string            `json:"resourceId"`   // the ID of the AWS resource
	Tags         map[string]string `json:"tags"`         // AWS tags
	Changes      []Change          `json:"changes"`
}

// function must satisfy AWS lambda.Start parameter spec.  Good enough: https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html
func Handle(ctx context.Context, event AWSConfigEvent) (Output, error) {
	output := Output{
		AccountID:    event.ConfigurationItem.AwsAccountID,
		ChangeTime:   event.ConfigurationItem.ConfigurationItemCaptureTime, // from configurationItemCaptureTime // TODO: ok?
		Region:       event.ConfigurationItem.AwsRegion,
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
