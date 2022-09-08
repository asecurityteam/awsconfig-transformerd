package v1

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_subnetTransformer_Create(t *testing.T) {
	baseConfigItem := configurationItem{
		AWSAccountID:                 "123456789012",
		AWSRegion:                    "us-west-2",
		ConfigurationItemCaptureTime: "2022-09-01T01:00:50.542Z",
		ResourceType:                 "AWS::EC2::Subnet",
		ARN:                          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
		Configuration:                json.RawMessage(`{"cidrBlock": "10.0.0.0/24"}`),
		Tags:                         map[string]string{"key1": "1"},
	}

	tests := []struct {
		name        string
		event       awsConfigEvent
		wantOutput  Output
		wantReject  bool
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful create",
			event: awsConfigEvent{
				ConfigurationItem: baseConfigItem,
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType:        "CREATE",
					ChangedProperties: map[string]json.RawMessage{},
				},
			},
			wantOutput: Output{
				AccountID:    "123456789012",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
				ResourceType: "AWS::EC2::Subnet",
				Region:       "us-west-2",
				ChangeTime:   "2022-09-01T01:00:50.542Z",
				Changes: []Change{
					{
						ChangeType: "ADDED",
						CIDRBlock:  "10.0.0.0/24",
					},
				},
				Tags: map[string]string{"key1": "1"},
			},
			wantReject: false,
			wantErr:    false,
		},
		{
			name: "malformed event - causes json.Unmarshal() to fail",
			event: awsConfigEvent{
				ConfigurationItem: configurationItem{
					AWSAccountID:                 "123456789012",
					AWSRegion:                    "us-west-2",
					ConfigurationItemCaptureTime: "2022-09-01T01:00:50.542Z",
					ResourceType:                 "AWS::EC2::Subnet",
					ARN:                          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
					Tags:                         map[string]string{"key1": "1"},
				},
			},
			wantOutput:  Output{},
			wantReject:  false,
			wantErr:     true,
			expectedErr: "unexpected end of JSON input",
		},
		{
			name: "missing account ID (first field checked)",
			event: awsConfigEvent{
				ConfigurationItem:     configurationItem{},
				ConfigurationItemDiff: configurationItemDiff{},
			},
			wantOutput:  Output{},
			wantReject:  false,
			wantErr:     true,
			expectedErr: ErrMissingValue{Field: "AWSAccountID"}.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := subnetTransformer{}
			gotOutput, gotReject, err := tr.Create(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("subnetTransformer.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expectedErr != "" {
				assert.Equal(t, tt.expectedErr, err.Error())
			}
			if !reflect.DeepEqual(gotOutput, tt.wantOutput) {
				t.Errorf("subnetTransformer.Create() gotOutput = %v, wantOutput %v", gotOutput, tt.wantOutput)
			}
			if gotReject != tt.wantReject {
				t.Errorf("subnetTransformer.Create() gotReject = %v, wantReject %v", gotReject, tt.wantReject)
			}
		})
	}
}

func Test_subnetTransformer_Update(t *testing.T) {
	baseConfigItem := configurationItem{
		AWSAccountID:                 "123456789012",
		AWSRegion:                    "us-west-2",
		ConfigurationItemCaptureTime: "2022-09-01T01:00:50.542Z",
		ResourceType:                 "AWS::EC2::Subnet",
		ARN:                          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
		Configuration:                json.RawMessage(`{"cidrBlock": "10.0.0.0/24"}`),
		Tags:                         map[string]string{"key1": "1"},
	}

	tests := []struct {
		name       string
		event      awsConfigEvent
		wantOutput Output
		wantReject bool
		wantErr    bool
	}{
		{
			name: "successful update",
			event: awsConfigEvent{
				ConfigurationItem: baseConfigItem,
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType:        "UPDATE",
					ChangedProperties: map[string]json.RawMessage{},
				},
			},
			wantOutput: Output{
				AccountID:    "123456789012",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
				ResourceType: "AWS::EC2::Subnet",
				Region:       "us-west-2",
				ChangeTime:   "2022-09-01T01:00:50.542Z",
				Tags:         map[string]string{"key1": "1"},
			},
			wantReject: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := subnetTransformer{}
			gotOutput, gotReject, err := tr.Update(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("subnetTransformer.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOutput, tt.wantOutput) {
				t.Errorf("subnetTransformer.Update() gotOutput = %v, wantOutput %v", gotOutput, tt.wantOutput)
			}
			if gotReject != tt.wantReject {
				t.Errorf("subnetTransformer.Update() gotReject = %v, want %v", gotReject, tt.wantReject)
			}
		})
	}
}

func Test_subnetTransformer_Delete(t *testing.T) {
	baseConfigItem := configurationItem{
		AWSAccountID:                 "123456789012",
		AWSRegion:                    "us-west-2",
		ConfigurationItemCaptureTime: "2022-09-01T01:00:50.542Z",
		ResourceType:                 "AWS::EC2::Subnet",
		ARN:                          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
		Configuration:                json.RawMessage(`{"cidrBlock": "10.0.0.0/24"}`),
	}

	tests := []struct {
		name        string
		event       awsConfigEvent
		wantOutput  Output
		wantReject  bool
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful delete",
			event: awsConfigEvent{
				ConfigurationItem: baseConfigItem,
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: "DELETE",
					ChangedProperties: map[string]json.RawMessage{
						"Configuration": json.RawMessage("{\"previousValue\":{\"cidrBlock\": \"10.0.0.0/24\"},\"updatedValue\":null,\"changeType\":\"DELETE\"}"),
					},
				},
			},
			wantOutput: Output{
				AccountID:    "123456789012",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
				ResourceType: "AWS::EC2::Subnet",
				Region:       "us-west-2",
				ChangeTime:   "2022-09-01T01:00:50.542Z",
				Changes: []Change{
					{
						ChangeType: "DELETED",
						CIDRBlock:  "10.0.0.0/24",
					},
				},
			},
			wantReject: false,
			wantErr:    false,
		},
		{
			name: "missing account ID (first field checked)",
			event: awsConfigEvent{
				ConfigurationItem:     configurationItem{},
				ConfigurationItemDiff: configurationItemDiff{},
			},
			wantOutput:  Output{},
			wantReject:  false,
			wantErr:     true,
			expectedErr: ErrMissingValue{Field: "AWSAccountID"}.Error(),
		},
		{
			name: "malformed - missing config diff",
			event: awsConfigEvent{
				ConfigurationItem: baseConfigItem,
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType:        "DELETE",
					ChangedProperties: map[string]json.RawMessage{},
				},
			},
			wantOutput:  Output{},
			wantReject:  false,
			wantErr:     true,
			expectedErr: "invalid configuration diff",
		},
		{
			name: "malformed - invalid config diff",
			event: awsConfigEvent{
				ConfigurationItem: baseConfigItem,
				ConfigurationItemDiff: configurationItemDiff{
					ChangeType: "DELETE",
					ChangedProperties: map[string]json.RawMessage{
						"Configuration": json.RawMessage("{\"previousValue\":{\"cidrBlock\": \"10.0.0.0/24\"}"),
					},
				},
			},
			wantOutput:  Output{},
			wantReject:  false,
			wantErr:     true,
			expectedErr: "unexpected end of JSON input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := subnetTransformer{}
			gotOutput, gotReject, err := tr.Delete(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("subnetTransformer.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expectedErr != "" {
				assert.Equal(t, tt.expectedErr, err.Error())
			}
			if !reflect.DeepEqual(gotOutput, tt.wantOutput) {
				t.Errorf("subnetTransformer.Delete() gotOutput = %v, wantOutput %v", gotOutput, tt.wantOutput)
			}
			if gotReject != tt.wantReject {
				t.Errorf("subnetTransformer.Delete() gotReject = %v, wantReject %v", gotReject, tt.wantReject)
			}
		})
	}
}

func Test_extractSubnetInfo(t *testing.T) {
	subnetConfig := subnetConfiguration{CIDRBlock: "10.0.0.0/24"}
	change := extractSubnetInfo(&subnetConfig)
	assert.Equal(t, change, Change{CIDRBlock: "10.0.0.0/24"})
}
