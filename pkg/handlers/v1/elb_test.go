package v1

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformELB(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
		ExpectError    bool
	}{
		{
			Name:      "elb-created",
			InputFile: "elb.create.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:06:49.363Z",
				Region:       "us-west-2",
				ResourceType: "AWS::ElasticLoadBalancing::LoadBalancer",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/config-test-elb",
				Tags: map[string]string{
					"key1": "1",
				},
				Changes: []Change{
					{
						Hostnames:  []string{"internal-config-test-elb-01234567.us-west-2.elb.amazonaws.com"},
						ChangeType: added,
					},
				},
			},
		},
		{
			Name:      "elb-updated",
			InputFile: "elb.update.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:12:28.624Z",
				Region:       "us-west-2",
				ResourceType: "AWS::ElasticLoadBalancing::LoadBalancer",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/config-test-elb",
				Tags: map[string]string{
					"key1": "1",
					"key2": "2",
				},
			},
		},
		{
			Name:      "elb-deleted",
			InputFile: "elb.delete.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:16:23.926Z",
				Region:       "us-west-2",
				ResourceType: "AWS::ElasticLoadBalancing::LoadBalancer",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/config-test-elb",
				Tags: map[string]string{
					"key1": "1",
					"key2": "2",
				},
				Changes: []Change{
					{
						Hostnames:  []string{"internal-config-test-elb-01234567.us-west-2.elb.amazonaws.com"},
						ChangeType: deleted,
					},
				},
			},
		},
		{
			Name:      "elbv2-created",
			InputFile: "elbv2.create.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:08:40.855Z",
				Region:       "us-west-2",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Tags: map[string]string{
					"key1": "1",
				},
				Changes: []Change{
					{
						Hostnames:  []string{"internal-config-test-alb-012345678.us-west-2.elb.amazonaws.com"},
						ChangeType: added,
					},
				},
			},
		},
		{
			Name:      "elbv2-updated",
			InputFile: "elbv2.update.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:12:03.211Z",
				Region:       "us-west-2",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Tags: map[string]string{
					"key1": "1",
					"key2": "2",
				},
			},
		},
		{
			Name:      "elbv2-deleted",
			InputFile: "elbv2.delete.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:16:22.178Z",
				Region:       "us-west-2",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Tags: map[string]string{
					"key1": "1",
					"key2": "2",
				},
				Changes: []Change{
					{
						Hostnames:  []string{"internal-config-test-alb-012345678.us-west-2.elb.amazonaws.com"},
						ChangeType: deleted,
					},
				},
			},
		},
		{
			Name:      "elbv2-created-notags",
			InputFile: "elbv2.create-notags.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-03-27T19:08:40.855Z",
				Region:       "us-west-2",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Changes: []Change{
					{
						Hostnames:  []string{"internal-config-test-alb-012345678.us-west-2.elb.amazonaws.com"},
						ChangeType: added,
					},
				},
			},
		},
		{
			Name:           "elb-malformed",
			InputFile:      "elb.malformed.json",
			ExpectedOutput: Output{},
			ExpectError:    true,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join("testdata", tt.InputFile))
			require.Nil(t, err)

			var input Input
			err = json.Unmarshal(data, &input)
			require.Nil(t, err)

			transformer := &Transformer{LogFn: logFn}
			output, err := transformer.Handle(context.Background(), input)
			if tt.ExpectError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ARN, output.ARN)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}

func TestELBTransformerCreate(t *testing.T) {
	tc := []struct {
		Name           string
		Event          awsConfigEvent
		ExpectedOutput Output
		ExpectError    bool
		ExpectedError  error
	}{
		{
			Name: "elb-unmarshall-error",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{"foo": "bar"},
					Configuration:                json.RawMessage(`{"dnsname": 1}`),
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: create,
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  &json.UnmarshalTypeError{Value: "number", Offset: 13, Type: reflect.TypeOf(""), Struct: "elbConfiguration", Field: "dnsname"},
		},
		{
			Name: "elb-missing-value",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: create,
					ChangedProperties: map[string]json.RawMessage{
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":null,\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  ErrMissingValue{Field: "AWSAccountID"},
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			et := elbTransformer{}
			output, err := et.Create(tt.Event)
			if tt.ExpectError {
				require.NotNil(t, err)
				assert.Equal(t, tt.ExpectedError, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ARN, output.ARN)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}

func TestELBTransformerUpdate(t *testing.T) {
	tc := []struct {
		Name           string
		Event          awsConfigEvent
		ExpectedOutput Output
		ExpectError    bool
		ExpectedError  error
	}{
		{
			Name: "elb-missing-value",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: update,
					ChangedProperties: map[string]json.RawMessage{
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":null,\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  ErrMissingValue{Field: "AWSAccountID"},
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			et := elbTransformer{}
			output, err := et.Update(tt.Event)
			if tt.ExpectError {
				require.NotNil(t, err)
				assert.Equal(t, tt.ExpectedError, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ARN, output.ARN)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}

func TestELBTransformerDelete(t *testing.T) {
	tc := []struct {
		Name           string
		Event          awsConfigEvent
		ExpectedOutput Output
		ExpectError    bool
		ExpectedError  error
	}{
		{
			Name: "elb-delete-no-tags",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"Configuration":                   json.RawMessage("{\"previousValue\":{\"loadBalancerName\":\"config-test-elb\",\"canonicalHostedZoneNameID\":\"Z1H1FL5HABSF5\",\"listenerDescriptions\":[{\"listener\":{\"protocol\":\"HTTP\",\"loadBalancerPort\":80,\"instanceProtocol\":\"HTTP\",\"instancePort\":80},\"policyNames\":[]}],\"policies\":{\"appCookieStickinessPolicies\":[],\"otherPolicies\":[],\"lbcookieStickinessPolicies\":[]},\"backendServerDescriptions\":[],\"availabilityZones\":[\"us-west-2a\",\"us-west-2b\",\"us-west-2c\",\"us-west-2d\"],\"subnets\":[\"subnet-24b88c41\",\"subnet-7b600d22\",\"subnet-94bbf1e3\",\"subnet-ee4140c6\"],\"sourceSecurityGroup\":{\"ownerAlias\":\"917546781095\",\"groupName\":\"default\"},\"securityGroups\":[\"sg-51164235\"],\"createdTime\":1553713467830,\"scheme\":\"internal\",\"dnsname\":\"internal-config-test-elb-67410663.us-west-2.elb.amazonaws.com\",\"vpcid\":\"vpc-8af6d7ef\"},\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":null,\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Region:       "us-west-2",
				ChangeTime:   "2019-03-27T19:06:49.363Z",
				Changes: []Change{
					{
						ChangeType: "DELETED",
						Hostnames:  []string{"internal-config-test-elb-67410663.us-west-2.elb.amazonaws.com"},
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "elb-delete-missing-supplementary-config",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"Configuration": json.RawMessage("{\"previousValue\":{\"loadBalancerName\":\"config-test-elb\",\"canonicalHostedZoneNameID\":\"Z1H1FL5HABSF5\",\"listenerDescriptions\":[{\"listener\":{\"protocol\":\"HTTP\",\"loadBalancerPort\":80,\"instanceProtocol\":\"HTTP\",\"instancePort\":80},\"policyNames\":[]}],\"policies\":{\"appCookieStickinessPolicies\":[],\"otherPolicies\":[],\"lbcookieStickinessPolicies\":[]},\"backendServerDescriptions\":[],\"availabilityZones\":[\"us-west-2a\",\"us-west-2b\",\"us-west-2c\",\"us-west-2d\"],\"subnets\":[\"subnet-24b88c41\",\"subnet-7b600d22\",\"subnet-94bbf1e3\",\"subnet-ee4140c6\"],\"sourceSecurityGroup\":{\"ownerAlias\":\"917546781095\",\"groupName\":\"default\"},\"securityGroups\":[\"sg-51164235\"],\"createdTime\":1553713467830,\"scheme\":\"internal\",\"dnsname\":\"internal-config-test-elb-67410663.us-west-2.elb.amazonaws.com\",\"vpcid\":\"vpc-8af6d7ef\"},\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ARN:          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
				ResourceType: "AWS::ElasticLoadBalancingV2::LoadBalancer",
				Region:       "us-west-2",
				ChangeTime:   "2019-03-27T19:06:49.363Z",
				Changes: []Change{
					{
						ChangeType: "DELETED",
						Hostnames:  []string{"internal-config-test-elb-67410663.us-west-2.elb.amazonaws.com"},
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "elb-delete-missing-config",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":null,\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  ErrMissingValue{Field: "ChangedProperties.Configuration"},
		},
		{
			Name: "elb-missing-value",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":null,\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  ErrMissingValue{Field: "AWSAccountID"},
		},
		{
			Name: "elb-unmarsall-error",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{"foo": "bar"},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"Configuration": json.RawMessage(`{"PreviousValue": 1}`),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  &json.UnmarshalTypeError{Value: "number", Offset: 19, Type: reflect.TypeOf(elbConfiguration{}), Struct: "elbConfigurationDiff", Field: "previousValue"},
		},
		{
			Name: "elb-unmarsall-supplementary-config-error",
			Event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2019-03-27T19:06:49.363Z",
					ResourceType:                 "AWS::ElasticLoadBalancingV2::LoadBalancer",
					ARN:                          "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/config-test-alb/5be197427c282f61",
					Tags:                         map[string]string{"foo": "bar"},
				},
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: delete,
					ChangedProperties: map[string]json.RawMessage{
						"Configuration":                   json.RawMessage("{\"previousValue\":{\"loadBalancerName\":\"config-test-elb\",\"canonicalHostedZoneNameID\":\"Z1H1FL5HABSF5\",\"listenerDescriptions\":[{\"listener\":{\"protocol\":\"HTTP\",\"loadBalancerPort\":80,\"instanceProtocol\":\"HTTP\",\"instancePort\":80},\"policyNames\":[]}],\"policies\":{\"appCookieStickinessPolicies\":[],\"otherPolicies\":[],\"lbcookieStickinessPolicies\":[]},\"backendServerDescriptions\":[],\"availabilityZones\":[\"us-west-2a\",\"us-west-2b\",\"us-west-2c\",\"us-west-2d\"],\"subnets\":[\"subnet-24b88c41\",\"subnet-7b600d22\",\"subnet-94bbf1e3\",\"subnet-ee4140c6\"],\"sourceSecurityGroup\":{\"ownerAlias\":\"917546781095\",\"groupName\":\"default\"},\"securityGroups\":[\"sg-51164235\"],\"createdTime\":1553713467830,\"scheme\":\"internal\",\"dnsname\":\"internal-config-test-elb-67410663.us-west-2.elb.amazonaws.com\",\"vpcid\":\"vpc-8af6d7ef\"},\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
						"SupplementaryConfiguration.Tags": json.RawMessage("{\"previousValue\":[{\"key\": 1}],\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			ExpectedOutput: Output{},
			ExpectError:    true,
			ExpectedError:  &json.UnmarshalTypeError{Value: "number", Offset: 27, Type: reflect.TypeOf(""), Struct: "tag", Field: "previousValue.key"},
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			et := elbTransformer{}
			output, err := et.Delete(tt.Event)
			if tt.ExpectError {
				require.NotNil(t, err)
				assert.Equal(t, tt.ExpectedError, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}
