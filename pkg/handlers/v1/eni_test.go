package v1

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTransformENI(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
		ExpectError    bool
	}{
		{
			Name:		"eni-created",
			InputFile:	"eni.create.json",
			ExpectedOutput: Output{
				AccountID:		"123456789",
				ChangeTime: 	"2020-08-21T12:00:00.000Z",
				Region:  		"ap-southeast-2",
				ResourceType:  	"AWS::EC2::NetworkInterface",
				ARN:			"arn:aws:ec2:ap-southeast-2:123456789:network-interface/eni-abcdefghi1234567",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.111.2.333"},
						RelatedResource: 	[]string{"micros-sec-example-ELB-AAAAAA11111"},
						ChangeType: 		create,
					},
				},
			},
		},
		{
			Name:  		"eni-deleted",
			InputFile:  "eni.delete.json",
			ExpectedOutput: Output{
				AccountID:  	"12345678910",
				ChangeTime:  	"2020-08-21T13:02:00.000Z",
				Region:  		"us-east-1",
				ResourceType:	"AWS::EC2::NetworkInterface",
				ARN:			"arn:aws:ec2:us-east-1:12345678910:network-interface/eni-hhhhhhh888888",
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.111.222.333"},
						RelatedResource: 	[]string{"app/marketp-ALB-eeeeeee5555555/ffffffff66666666"},
						ChangeType: 		delete,
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

func TestENITransformerUpdate(t *testing.T){
	// FOr now it looks like updates don't apply to the types of ENI events we're interested in
	data, err := ioutil.ReadFile(filepath.Join("testdata", "eni.update.json"))
	require.Nil(t, err)

	var input Input
	err = json.Unmarshal(data, &input)
	require.Nil(t, err)

	transformer := &Transformer{LogFn: logFn}
	output, err := transformer.Handle(context.Background(), input)
	require.Nil(t, err)

	emptyOutput := Output{}

	assert.True(t,reflect.DeepEqual(output, emptyOutput), "Expected an empty Output")
}