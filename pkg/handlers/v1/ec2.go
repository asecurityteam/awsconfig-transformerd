package v1

import (
	"time"
)

type ec2Configuration struct {
	InstanceID string `json:"instanceId"`
	State      struct {
		Code int    `json:"code"`
		Name string `json:"name"`
	} `json:"state"`
	StateTransitionReason string                               `json:"stateTransitionReason"`
	InstanceType          string                               `json:"instanceType"`
	LaunchTime            time.Time                            `json:"launchTime"`
	NetworkInterfaces     []configurationNetworkInterfaceValue `json:"networkInterfaces"`
}

type configurationNetworkInterfaceValue struct {
	NetworkInterfaceID string `json:"networkInterfaceId"`
	SubnetID           string `json:"subnetId"`
	VpcID              string `json:"vpcId"`
	Description        string `json:"description"`
	OwnerID            string `json:"ownerId"`
	PrivateDNSName     string `json:"privateDnsName"`
	Association        struct {
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

type configurationNetworkInterface struct {
	PreviousValue *configurationNetworkInterfaceValue `json:"previousValue"`
	UpdatedValue  *configurationNetworkInterfaceValue `json:"updatedValue"`
	ChangeType    string                              `json:"changeType"`
}
