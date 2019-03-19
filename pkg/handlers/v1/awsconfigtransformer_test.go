package v1

import (
	"bitbucket.org/asecurityteam/awsconfig-transformerd/pkg/domain"
	"context"
	"encoding/json"
	"github.com/asecurityteam/logevent"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSConfigEventMarshalling(t *testing.T) {

	const filename = "awsconfigpayload.json"

	str, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read file '%s': %s", filename, err)
	}

	// if you want to see the raw string:
	// fmt.Println(string(str))

	res := awsConfigEvent{}
	json.Unmarshal(str, &res)

	assert.NotNil(t, res.ConfigurationItemDiff.ChangedProperties, "marshalling should have resulted in non-nil value")
	marshalled, _ := json.MarshalIndent(res, "", "    ")
	// if you want to see it:
	// fmt.Println(string(marshalled))

	jsonEquals, _ := jsonBytesEqual([]byte(str), marshalled)

	assert.True(t, jsonEquals, "expect exact JSON equality before marshall/unmarshal and after.  "+
		"If they're not, it's probably because your JSON key does not match the field name "+
		"or you did not use a pointer where you should have")

}

func TestTransformEmptiness(t *testing.T) {
	event := awsConfigEvent{}
	marshalled, _ := json.Marshal(event)
	transformer := &Transformer{LogFn: func(_ context.Context) domain.Logger {
		return logevent.New(logevent.Config{Output: ioutil.Discard})
	}}
	_, err := transformer.Handle(context.Background(), Input{Message: string(marshalled)})
	assert.Nil(t, err, "expected non-nil")
}

func TestTransformInvalidJSON(t *testing.T) {
	transformer := &Transformer{}
	_, err := transformer.Handle(context.Background(), Input{Message: "not json"})
	assert.NotNil(t, err, "expected non-nil")
}

func TestTransformEC2(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
	}{
		{
			Name:      "ec2-created",
			InputFile: "ec2.0.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T20:43:10.208Z",
				Region:       "us-west-2",
				ResourceID:   "i-0a763ac3ee37d8d2b",
				ResourceType: "AWS::EC2::Instance",
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
				ResourceID:   "i-0a763ac3ee37d8d2b",
				ResourceType: "AWS::EC2::Instance",
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
				ResourceID:   "i-0a763ac3ee37d8d2b",
				ResourceType: "AWS::EC2::Instance",
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
				ResourceID:   "i-0a763ac3ee37d8d2b",
				ResourceType: "AWS::EC2::Instance",
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
				ResourceID:   "i-0a763ac3ee37d8d2b",
				ResourceType: "AWS::EC2::Instance",
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
	}

	for _, tt := range tc {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join("testdata", tt.InputFile))
			require.Nil(t, err)

			var input Input
			err = json.Unmarshal(data, &input)
			require.Nil(t, err)

			transformer := &Transformer{}
			output, err := transformer.Handle(context.Background(), input)
			require.Nil(t, err)

			assert.Equal(t, tt.ExpectedOutput.AccountID, output.AccountID)
			assert.Equal(t, tt.ExpectedOutput.Region, output.Region)
			assert.Equal(t, tt.ExpectedOutput.ResourceID, output.ResourceID)
			assert.Equal(t, tt.ExpectedOutput.ResourceType, output.ResourceType)
			assert.Equal(t, tt.ExpectedOutput.Tags, output.Tags)
			assert.Equal(t, tt.ExpectedOutput.ChangeTime, output.ChangeTime)
			assert.True(t, reflect.DeepEqual(tt.ExpectedOutput.Changes, output.Changes), "The expected changes were different than the result")
		})
	}
}

// JSONBytesEqual compares the JSON in two byte slices.
func jsonBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
