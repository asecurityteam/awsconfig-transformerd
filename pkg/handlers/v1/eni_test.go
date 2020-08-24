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

func TestTransformENI(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
		ExpectError    bool
	}{
		{
			Name:      "eni-created",
			InputFile: "eni.create.json",
			ExpectedOutput: Output{
				AccountID:    "123456789",
				ChangeTime:   "2020-08-21T12:00:00.000Z",
				Region:       "ap-southeast-2",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:ap-southeast-2:123456789:network-interface/eni-abcdefghi1234567",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.111.222.333"},
						RelatedResources:   []string{"micros-sec-example-ELB-AAAAAA11111"},
						ChangeType:         added,
					},
				},
			},
		},
		{
			Name:      "eni-updated",
			InputFile: "eni.update.json",
			ExpectedOutput: Output{
				AccountID:    "12345678910",
				ChangeTime:   "2020-08-21T12:31:00.000Z",
				Region:       "eu-central-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:eu-central-1:12345678910:network-interface/eni-eeeeeee8888888",
			},
		},
		{
			Name:      "eni-deleted",
			InputFile: "eni.delete.json",
			ExpectedOutput: Output{
				AccountID:    "12345678910",
				ChangeTime:   "2020-08-21T13:02:00.000Z",
				Region:       "us-east-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:us-east-1:12345678910:network-interface/eni-hhhhhhh888888",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.111.222.333"},
						RelatedResources:   []string{"app/marketp-ALB-eeeeeee5555555/ffffffff66666666"},
						ChangeType:         deleted,
					},
				},
			},
		},
	}

	// TODO: This is shared with at least ELB tests and maybe EC2. Pull out to own testing function?
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
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}

func TestFilterENI(t *testing.T) {
	filteredConfig := eniConfiguration{
		Description:        "ELB app/never-used",
		PrivateIPAddresses: nil,
		RequesterID:        elbManaged,
		RequesterManaged:   false,
	}
	jsonFilteredConfig, err := json.Marshal(filteredConfig)
	if err != nil {
		print("Could not marshal test struct")
	}

	filteredConfigItem := configurationItem{
		Configuration:                json.RawMessage(jsonFilteredConfig),
		AWSAccountID:                 "123456789",
		ResourceType:                 "AWS::EC2::NetworkInterface",
		ARN:                          "arn:aws:ec2:us-east-1:12345678910:network-interface/eni-hhhhhhh888888",
		AWSRegion:                    "us-east-1",
		ConfigurationItemCaptureTime: "2020-08-21T13:00:01.000Z",
	}

	filteredConfigEvent := awsConfigEvent{
		ConfigurationItem: filteredConfigItem,
	}

	transformer := eniTransformer{}
	emptyOutput := Output{}

	createOutput, err := transformer.Create(filteredConfigEvent)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(createOutput, emptyOutput), "Expected empty output due to filtering")

	updateOutput, err := transformer.Update(filteredConfigEvent)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(updateOutput, emptyOutput), "Expected empty output due to filtering")

	// Because deleting looks at the PreviousValue for filtering, the previous configuration does not matter
	filteredConfigDiff := eniConfigurationDiff{PreviousValue: &filteredConfig}
	jsonFilteredConfigDiff, err := json.Marshal(filteredConfigDiff)
	if err != nil {
		print("Could not marshal test struct")
	}

	filteredConfigEvent.ConfigurationItemDiff = configurationItemDiff{
		ChangedProperties: map[string]json.RawMessage{
			"Configuration": json.RawMessage(jsonFilteredConfigDiff),
		},
	}

	deleteOutput, err := transformer.Delete(filteredConfigEvent)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(deleteOutput, emptyOutput), "Expected empty output due to filtering")
}
