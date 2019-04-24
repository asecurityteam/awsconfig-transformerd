package v1

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/asecurityteam/awsconfig-transformerd/pkg/domain"
	"github.com/asecurityteam/logevent"
	"github.com/asecurityteam/runhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func logFn(_ context.Context) domain.Logger {
	return logevent.New(logevent.Config{Output: ioutil.Discard})
}

func TestAWSConfigEventMarshalling(t *testing.T) {

	const filename = "awsconfigpayload.json"

	data, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read file '%s': %s", filename, err)
	}

	res := awsConfigEvent{}
	_ = json.Unmarshal(data, &res)

	assert.NotNil(t, res.ConfigurationItemDiff.ChangedProperties, "marshalling should have resulted in non-nil value")
	marshaled, _ := json.MarshalIndent(res, "", "    ")

	assert.JSONEq(t, string(data), string(marshaled))
}

func TestTransformEmptiness(t *testing.T) {
	event := awsConfigEvent{}
	marshalled, _ := json.Marshal(event)
	transformer := &Transformer{LogFn: logFn}
	output, err := transformer.Handle(context.Background(), Input{Message: string(marshalled)})
	assert.Nil(t, err, "expected non-nil")
	assert.Equal(t, 0, len(output.Changes))
}

func TestTransformInvalidJSON(t *testing.T) {
	transformer := &Transformer{LogFn: logFn}
	_, err := transformer.Handle(context.Background(), Input{Message: "not json"})
	assert.NotNil(t, err, "expected non-nil")
}

func TestTransformEC2(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
		ExpectError    bool
	}{
		{
			Name:      "ec2-created",
			InputFile: "ec2.0.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T20:43:10.208Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
				Tags: map[string]string{
					"business_unit": "CISO-Security",
					"service_name":  "foo-bar",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"172.31.30.79"},
						PublicIPAddresses:  []string{"34.222.120.66"},
						Hostnames:          []string{"ec2-34-222-120-66.us-west-2.compute.amazonaws.com"},
						ChangeType:         "ADDED",
					},
				},
			},
		},
		{
			Name:      "ec2-stopped",
			InputFile: "ec2.1.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T20:48:32.538Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
				Tags: map[string]string{
					"business_unit": "CISO-Security",
					"service_name":  "foo-bar",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{},
						PublicIPAddresses:  []string{"34.222.120.66"},
						Hostnames:          []string{"ec2-34-222-120-66.us-west-2.compute.amazonaws.com"},
						ChangeType:         "DELETED",
					},
				},
			},
		},
		{
			Name:      "ec2-restarted",
			InputFile: "ec2.2.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T21:02:18.758Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
				Tags: map[string]string{
					"business_unit": "CISO-Security",
					"service_name":  "foo-bar",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{},
						PublicIPAddresses:  []string{"34.219.72.29"},
						Hostnames:          []string{"ec2-34-219-72-29.us-west-2.compute.amazonaws.com"},
						ChangeType:         "ADDED",
					},
				},
			},
		},
		{
			Name:      "ec2-stopped-again",
			InputFile: "ec2.3.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T21:17:53.073Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
				Tags: map[string]string{
					"business_unit": "CISO-Security",
					"service_name":  "foo-bar",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{},
						PublicIPAddresses:  []string{"34.219.72.29"},
						Hostnames:          []string{"ec2-34-219-72-29.us-west-2.compute.amazonaws.com"},
						ChangeType:         "DELETED",
					},
				},
			},
		},
		{
			Name:      "ec2-terminated",
			InputFile: "ec2.4.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T21:31:57.042Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
				Tags: map[string]string{
					"business_unit": "CISO-Security",
					"service_name":  "foo-bar",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"172.31.30.79"},
						ChangeType:         "DELETED",
					},
				},
			},
		},
		{
			Name:        "ec2-malformed-configuration",
			InputFile:   "ec2.malformed.json",
			ExpectError: true,
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

			transformer := &Transformer{StatFn: runhttp.StatFromContext, LogFn: logFn}
			output, err := transformer.Handle(context.Background(), input)
			if tt.ExpectError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.ARN, output.ARN)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}
