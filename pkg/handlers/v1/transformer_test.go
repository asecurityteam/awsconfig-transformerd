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

	assert.NotNil(t, res.ConfigurationItemDiff.ChangedProperties, "marshaling should have resulted in non-nil value")
	marshaled, _ := json.MarshalIndent(res, "", "    ")

	assert.JSONEq(t, string(data), string(marshaled))
}

func TestTransformEmptiness(t *testing.T) {
	event := awsConfigEvent{}
	marshaled, _ := json.Marshal(event)
	transformer := &Transformer{LogFn: logFn}
	output, err := transformer.Handle(context.Background(), Input{Message: string(marshaled)})
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
			Name:      "ec2-created-notags",
			InputFile: "ec2.5.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T20:43:10.208Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
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
			Name:      "ec2-terminated-notags",
			InputFile: "ec2.6.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2019-02-22T21:31:57.042Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:instance/i-0a763ac3ee37d8d2b",
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
		{
			Name:      "ec2-deleted-configuration",
			InputFile: "ec2.deleted.json",
			ExpectedOutput: Output{
				AccountID:    "752631980301",
				ChangeTime:   "2019-12-11T01:00:29.000Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Instance",
				ARN:          "arn:aws:ec2:us-west-2:752631980301:instance/i-08f37101ae44e31e4",
				Tags: map[string]string{
					"aws:autoscaling:groupName":     "status-page-web-graphql--stg-west2--314741c3cad246cebf35d178c10d086750--2019-12-11-00-56-utc--ohdjsjstilmha6fg--WebServer",
					"aws:cloudformation:logical-id": "WebServer",
					"aws:cloudformation:stack-id":   "arn:aws:cloudformation:us-west-2:752631980301:stack/status-page-web-graphql--stg-west2--314741c3cad246cebf35d178c10d086750--2019-12-11-00-56-utc--ohdjsjstilmha6fg/28c7b740-1bb1-11ea-ae0e-024a7c148296",
					"aws:cloudformation:stack-name": "status-page-web-graphql--stg-west2--314741c3cad246cebf35d178c10d086750--2019-12-11-00-56-utc--ohdjsjstilmha6fg",
					"business_unit":                 "Engineering-SP",
					"Name":                          "status-page-web-graphql--stg-west2--314741c3cad246cebf35d178c10d086750--2019-12-11-00-56-utc--ohdjsjstilmha6fg",
					"chaos_monkey":                  "false",
					"compute_type":                  "ec2",
					"deployment_id":                 "ohdjsjstilmha6fg",
					"environment":                   "stg-west2",
					"environment_type":              "staging",
					"micros_deployment_id":          "ohdjsjstilmha6fg",
					"micros_group":                  "WebServer",
					"micros_service_id":             "status-page-web-graphql",
					"micros_service_version":        "314741c3cad246cebf35d178c10d08675093b3dc",
					"resource_owner":                "rvenkatesh",
					"service_name":                  "status-page-web-graphql.us-west-2.staging.atl-paas.net",
				},
				Changes: []Change{
					{
						PrivateIPAddresses: []string{"10.103.19.93", "10.107.70.212"},
						PublicIPAddresses:  []string{"52.27.166.73"},
						Hostnames:          []string{"ec2-52-27-166-73.us-west-2.compute.amazonaws.com"},
						ChangeType:         "DELETED",
					},
				},
			},
		},
		{
			Name:        "ec2-deleted-malformed-config",
			InputFile:   "ec2.deleted-malformed.json",
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
			assert.ElementsMatch(t, tt.ExpectedOutput.Changes, output.Changes)
		})
	}
}

func TestTransformSubnet(t *testing.T) {
	tc := []struct {
		Name           string
		InputFile      string
		ExpectedOutput Output
		ExpectError    bool
		ExpectedError  error
	}{
		{
			Name:      "subnet-created",
			InputFile: "subnet.create.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2022-09-01T01:00:50.542Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Subnet",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
				Tags: map[string]string{
					"key1": "1",
				},
				Changes: []Change{
					{
						CIDRBlock:  "10.0.0.0/24",
						ChangeType: added,
					},
				},
			},
			ExpectError: false,
		},
		{
			Name:      "subnet-deleted",
			InputFile: "subnet.delete.json",
			ExpectedOutput: Output{
				AccountID:    "123456789012",
				ChangeTime:   "2022-09-01T01:05:14.878Z",
				Region:       "us-west-2",
				ResourceType: "AWS::EC2::Subnet",
				ARN:          "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-000aa0a000a00a0aa",
				Tags:         map[string]string{},
				Changes: []Change{
					{
						CIDRBlock:  "10.0.0.0/24",
						ChangeType: deleted,
					},
				},
			},
			ExpectError: false,
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
			assert.ElementsMatch(t, tt.ExpectedOutput.Changes, output.Changes)
		})
	}
}

func Test_extractTagChanges(t *testing.T) {
	tests := []struct {
		name    string
		ev      configurationItemDiff
		want    []TagChange
		wantErr bool
	}{
		{
			"no tags",
			configurationItemDiff{
				ChangedProperties: map[string]json.RawMessage{
					"Configuration.NotTag.1":                json.RawMessage("Something"),
					"SupplementaryConfiguration.Over9000.1": json.RawMessage("does not matter"),
				},
				ChangeType: "ADD",
			},
			make([]TagChange, 0),
			false,
		},
		{
			"malformed",
			configurationItemDiff{
				ChangedProperties: map[string]json.RawMessage{
					"Configuration.TagSet.1": json.RawMessage("not JSON"),
				},
				ChangeType: "ADD",
			},
			make([]TagChange, 0),
			true,
		},
		{
			"both nil",
			configurationItemDiff{
				ChangedProperties: map[string]json.RawMessage{
					"Configuration.TagSet.1": json.RawMessage("{}"),
				},
				ChangeType: "ADD",
			},
			make([]TagChange, 0),
			true,
		},
		{
			"tags included",
			configurationItemDiff{
				ChangedProperties: map[string]json.RawMessage{
					"Configuration.Tags.1": json.RawMessage("{\"previousValue\": null, \"updatedValue\": {\"key\": \"info\", \"value\": \"I added a new tag\"}}"),
				},
				ChangeType: "ADD",
			},
			[]TagChange{
				{
					UpdatedValue: &Tag{
						Key:   "info",
						Value: "I added a new tag",
					},
					PreviousValue: nil,
				},
			},
			false,
		},
		// happy path is tested via regular add/del events for resources
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTagChanges(tt.ev)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTagChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTagChanges() got = %v, want %v", got, tt.want)
			}
		})
	}
}
