package v1

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/stretchr/testify/require"
)

func TestMissingRequiredFields(t *testing.T) {
	tc := []struct {
		Name          string
		ConfigItem    configurationItem
		ExpectedError bool
	}{
		{
			Name: "missing-accountID",
			ConfigItem: configurationItem{
				AWSAccountID:                 "",
				AWSRegion:                    "us-west-2",
				ConfigurationItemCaptureTime: "2019-02-22T20:19:20.543Z",
				ResourceID:                   "abc1234",
				ResourceType:                 configservice.ResourceTypeAwsEc2Instance,
			},
			ExpectedError: true,
		},
		{
			Name: "missing-region",
			ConfigItem: configurationItem{
				AWSAccountID:                 "0123456789012",
				AWSRegion:                    "",
				ConfigurationItemCaptureTime: "2019-02-22T20:19:20.543Z",
				ResourceID:                   "abc1234",
				ResourceType:                 configservice.ResourceTypeAwsEc2Instance,
			},
			ExpectedError: true,
		},
		{
			Name: "missing-time",
			ConfigItem: configurationItem{
				AWSAccountID:                 "0123456789012",
				AWSRegion:                    "us-west-2",
				ConfigurationItemCaptureTime: "",
				ResourceID:                   "abc1234",
				ResourceType:                 configservice.ResourceTypeAwsEc2Instance,
			},
			ExpectedError: true,
		},
		{
			Name: "missing-resource-type",
			ConfigItem: configurationItem{
				AWSAccountID:                 "0123456789012",
				AWSRegion:                    "us-west-2",
				ConfigurationItemCaptureTime: "2019-02-22T20:19:20.543Z",
				ResourceID:                   "abc1234",
				ResourceType:                 "",
			},
			ExpectedError: true,
		},
		{
			Name: "empty tags",
			ConfigItem: configurationItem{
				AWSAccountID:                 "0123456789012",
				AWSRegion:                    "us-west-2",
				ConfigurationItemCaptureTime: "2019-02-22T20:19:20.543Z",
				ResourceID:                   "abc1234",
				ARN:                          "arn:partition:service:region:account-id:resourcetype/resource",
				ResourceType:                 configservice.ResourceTypeAwsEc2Instance,
			},
			ExpectedError: false,
		},
		{
			Name: "valid",
			ConfigItem: configurationItem{
				AWSAccountID:                 "0123456789012",
				AWSRegion:                    "us-west-2",
				ConfigurationItemCaptureTime: "2019-02-22T20:19:20.543Z",
				ResourceID:                   "abc1234",
				ARN:                          "arn:partition:service:region:account-id:resourcetype/resource",
				ResourceType:                 configservice.ResourceTypeAwsEc2Instance,
				Tags:                         map[string]string{"foo": "bar"},
			},
			ExpectedError: false,
		},
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			_, err := getBaseOutput(tt.ConfigItem)
			if tt.ExpectedError {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}
