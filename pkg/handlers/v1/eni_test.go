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
			InputFile: "eni.1.create.json",
			ExpectedOutput: Output{
				AccountID:    "123456789123",
				ChangeTime:   "2020-08-21T12:00:00.000Z",
				Region:       "ap-southeast-2",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:ap-southeast-2:123456789:network-interface/eni-abcdefghi1234567",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.111.222.33"},
						RelatedResources:   []string{"micros-sec-example-ELB-AAAAAA11111"},
						ChangeType:         added,
					},
				},
			},
		},
		{
			Name:      "eni-createed-with-public-ip",
			InputFile: "eni.2.create.json",
			ExpectedOutput: Output{
				AccountID:    "000000000000",
				ChangeTime:   "2021-11-01T03:09:57.320Z",
				Region:       "eu-central-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:eu-central-1:000000000000:network-interface/eni-bbbbbbbb9999999",
				Changes: []Change{
					{
						PublicIPAddresses:  []string{"18.111.200.30"},
						PrivateIPAddresses: []string{"10.111.222.138"},
						Hostnames:          []string{"ec2-18-111-200-30.eu-central-1.compute.amazonaws.com"},
						RelatedResources:   []string{"app/ALB-1212412/12412413"},
						ChangeType:         added,
					},
				},
			},
		},
		{
			Name:      "eni-updated",
			InputFile: "eni.1.update.json",
			ExpectedOutput: Output{
				AccountID:    "112233445566",
				ChangeTime:   "2020-08-21T12:31:00.000Z",
				Region:       "eu-central-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:eu-central-1:112233445566:network-interface/eni-eeeeeee8888888",
			},
		},
		{
			Name:      "eni-updated-with-public-ip",
			InputFile: "eni.2.update.json",
			ExpectedOutput: Output{
				AccountID:    "123123123123",
				ChangeTime:   "2021-11-01T12:56:57.500Z",
				Region:       "us-west-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:us-west-1:123123123123:network-interface/eni-123456789",
				Changes: []Change{
					{
						PublicIPAddresses:  []string{"54.111.25.212"},
						PrivateIPAddresses: []string{},
						Hostnames:          []string{"ec2-54-111-25-212.us-west-1.compute.amazonaws.com"},
						RelatedResources:   []string{"micros-sec-example-ELB-BBBBBBBB222222"},
						ChangeType:         added,
					},
				},
			},
		},
		{
			Name:      "eni-updated-with-deleted-private-ip",
			InputFile: "eni.3.update.json",
			ExpectedOutput: Output{
				AccountID:    "010203040506",
				ChangeTime:   "2021-11-01T12:56:57.500Z",
				Region:       "us-west-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:us-west-1:010203040506:network-interface/eni-123456789",
				Changes: []Change{
					{
						PublicIPAddresses:  []string{},
						PrivateIPAddresses: []string{"10.23.24.25"},
						Hostnames:          []string{},
						RelatedResources:   []string{"micros-sec-example-ELB-AAAAAAAABBBBBBB111111"},
						ChangeType:         deleted,
					},
				},
			},
		},
		{
			Name:      "eni-deleted",
			InputFile: "eni.1.delete.json",
			ExpectedOutput: Output{
				AccountID:    "098765432109",
				ChangeTime:   "2020-08-21T13:02:00.000Z",
				Region:       "us-east-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:us-east-1:098765432109:network-interface/eni-hhhhhhh888888",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.11.22.33"},
						RelatedResources:   []string{"app/marketp-ALB-eeeeeee5555555/ffffffff66666666"},
						ChangeType:         deleted,
					},
				},
			},
		},
		{
			Name:      "eni-deleted-with-public-ip",
			InputFile: "eni.2.delete.json",
			ExpectedOutput: Output{
				AccountID:    "123456789123",
				ChangeTime:   "2021-11-05T00:00:07.736Z",
				Region:       "eu-central-1",
				ResourceType: "AWS::EC2::NetworkInterface",
				ARN:          "arn:aws:ec2:eu-central-1:123456789123:network-interface/eni-abcd1234",
				Changes: []Change{
					{
						PublicIPAddresses:  []string{"18.123.152.102"},
						PrivateIPAddresses: []string{"10.13.56.162"},
						Hostnames:          []string{"ec2-18-123-152-102.eu-central-1.compute.amazonaws.com"},
						RelatedResources:   []string{"app/marketp-ALB-fadg3t3t55555/gggggggg99999999"},
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
			assert.ElementsMatch(t, tt.ExpectedOutput.Changes, output.Changes)
		})
	}
}

func TestFilterENI(t *testing.T) {
	filteredConfig := eniConfiguration{
		Description:        "ELB app/never-used",
		PrivateIPAddresses: nil,
		RequesterID:        elbRequester,
		RequesterManaged:   false,
	}
	jsonFilteredConfig, err := json.Marshal(filteredConfig)
	if err != nil {
		print("Could not marshal test struct")
	}

	filteredConfigItem := configurationItem{
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

	t.Run("eni-created", func(t *testing.T) {
		filteredConfig.RequesterManaged = false
		filteredConfigEvent.ConfigurationItem.Configuration = json.RawMessage(jsonFilteredConfig)

		createOutput, reject, err := transformer.Create(filteredConfigEvent)
		assert.Equal(t, true, reject)
		assert.Nil(t, err)
		assert.True(t, createOutput.Changes == nil, "Expected empty changes due to filtering")
	})

	t.Run("eni-created-with-tags", func(t *testing.T) {
		filteredConfig.RequesterManaged = true
		filteredConfigEvent.ConfigurationItem.Configuration = json.RawMessage(jsonFilteredConfig)
		filteredConfigEvent.ConfigurationItemDiff = configurationItemDiff{
			ChangedProperties: map[string]json.RawMessage{
				"Configuration.NotTag.1": json.RawMessage("{\"previousValue\": null, \"updatedValue\": {\"key\": \"info\", \"value\": \"I added a new tag\"}}"),
			},
		}

		createOutput, reject, err := transformer.Create(filteredConfigEvent)
		assert.Equal(t, true, reject)
		assert.Nil(t, err)
		assert.True(t, createOutput.Changes == nil, "Expected empty changes due to filtering")
	})

	t.Run("eni-updated", func(t *testing.T) {
		filteredConfig.RequesterManaged = true
		filteredConfigDiff := eniConfigurationDiff{PreviousValue: &filteredConfig, UpdatedValue: &filteredConfig, ChangeType: update}
		jsonFilteredConfigDiff, err := json.Marshal(filteredConfigDiff)
		if err != nil {
			print("Could not marshal test struct")
		}

		filteredConfigEvent.ConfigurationItemDiff = configurationItemDiff{
			ChangedProperties: map[string]json.RawMessage{
				"Configuration":          json.RawMessage(jsonFilteredConfigDiff),
				"Configuration.NotTag.1": json.RawMessage("{\"previousValue\": null, \"updatedValue\": {\"key\": \"info\", \"value\": \"I added a new tag\"}}"),
			},
		}

		updateOutput, reject, err := transformer.Update(filteredConfigEvent)
		assert.Equal(t, true, reject)
		assert.Nil(t, err)
		assert.True(t, updateOutput.Changes == nil, "Expected empty changes due to filtering")
	})

	t.Run("eni-deleted", func(t *testing.T) {
		filteredConfig.RequesterManaged = false
		// Because deleting looks at the PreviousValue for filtering, we need some more set up
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

		deleteOutput, _, err := transformer.Delete(filteredConfigEvent)
		assert.Nil(t, err)
		assert.True(t, deleteOutput.Changes == nil, "Expected empty changes due to filtering")
	})
}

func TestErrorENI(t *testing.T) {
	malformedConfigItem := configurationItem{
		// excluding AWSAccountID so that we can cause a missingField error to bubble up from the transformer
		ResourceType:                 "AWS::EC2::NetworkInterface",
		ARN:                          "arn:aws:ec2:us-east-1:12345678910:network-interface/eni-hhhhhhh888888",
		AWSRegion:                    "us-east-1",
		ConfigurationItemCaptureTime: "2020-08-21T13:00:01.000Z",
	}

	malformedConfigEvent := awsConfigEvent{
		ConfigurationItem: malformedConfigItem,
	}

	transformer := eniTransformer{}

	t.Run("malformed-create-event", func(t *testing.T) {
		_, _, err := transformer.Create(malformedConfigEvent)
		assert.NotNil(t, err)
		assert.Equal(t, err, ErrMissingValue{Field: "AWSAccountID"})
	})

	t.Run("malformed-update-event", func(t *testing.T) {
		_, _, err := transformer.Update(malformedConfigEvent)
		assert.NotNil(t, err)
		assert.Equal(t, err, ErrMissingValue{Field: "AWSAccountID"})
	})

	t.Run("malformed-delete-event", func(t *testing.T) {
		_, _, err := transformer.Delete(malformedConfigEvent)
		assert.NotNil(t, err)
		assert.Equal(t, err, ErrMissingValue{Field: "AWSAccountID"})
	})

	// We would like this to pass evaluation so we can instead test unmarshaling errors for configs
	malformedConfigEvent.ConfigurationItem.AWSAccountID = "123456789"

	t.Run("malformed-json-update-event", func(t *testing.T) {
		malformedConfigEvent.ConfigurationItem.Configuration = []byte("{ bad: json }")
		_, _, err := transformer.Update(malformedConfigEvent)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})

	malformedConfiguration := json.RawMessage(`{"requesterManaged": "sure"}`)

	t.Run("malformed-create-config", func(t *testing.T) {
		malformedConfigEvent.ConfigurationItem.Configuration = malformedConfiguration
		_, _, err := transformer.Create(malformedConfigEvent)
		assert.NotNil(t, err)
		expected := &json.UnmarshalTypeError{Value: "string", Offset: 27, Type: reflect.TypeOf(false), Struct: "eniConfiguration", Field: "requesterManaged"}
		assert.Equal(t, expected, err)
	})

	t.Run("malformed-delete-previousConfig", func(t *testing.T) {
		malformedConfigEvent.ConfigurationItemDiff = configurationItemDiff{
			ChangedProperties: map[string]json.RawMessage{
				"Configuration": json.RawMessage(`{"previousValue": "bad"}`),
			},
		}
		_, _, err := transformer.Delete(malformedConfigEvent)
		assert.NotNil(t, err)
		expected := &json.UnmarshalTypeError{Value: "string", Offset: 23, Type: reflect.TypeOf(eniConfiguration{}), Struct: "eniConfigurationDiff", Field: "previousValue"}
		assert.Equal(t, expected, err)
	})

	//Edge case where generic config event fields look fine, but IP Block is malformed
	t.Run("malformed-private-ip-json-block-update-event", func(t *testing.T) {
		emptyIPBlockConfigurationItem := configurationItem{
			Configuration:                json.RawMessage(`{"description": "FILLER","privateIpAddresses": [],"requesterId": "amazon-elb","requesterManaged": true}`),
			ConfigurationItemCaptureTime: "2021-11-02T12:56:57.562Z",
			AWSAccountID:                 "111111111111",
			ResourceType:                 "AWS::EC2::NetworkInterface",
			ARN:                          "arn:aws:ec2:us-west-1:752631980301:network-interface/eni-0f0a311411ae5166d",
			AWSRegion:                    "us-west-1",
		}

		okayConfigurationItemDiff := configurationItemDiff{
			ChangedProperties: map[string]json.RawMessage{
				"Configuration.PrivateIpAddresses.0": json.RawMessage(`{"previousValue": "bad"}`),
			},
		}

		malformedEvent := awsConfigEvent{
			ConfigurationItemDiff: okayConfigurationItemDiff,
			ConfigurationItem:     emptyIPBlockConfigurationItem,
		}

		_, _, err := transformer.Update(malformedEvent)
		assert.NotNil(t, err)
		expected := &json.UnmarshalTypeError{Value: "string", Type: reflect.TypeOf(privateIPAddress{}), Offset: 23, Struct: "privateIPBlockDiff", Field: "previousValue"}
		assert.Equal(t, expected, err)
	})
}
