package v1

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshalling(t *testing.T) {

	const filename = "awsconfigpayload.json"

	str, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read file '%s': %s", filename, err)
	}

	// if you want to see the raw string:
	// fmt.Println(string(str))

	res := awsConfigEvent{}
	json.Unmarshal(str, &res)

	// spot check a few things to confirm both static marshalling ^^^ and dynamic (via get* functions) all work

	assert.NotNil(t, res.ConfigurationItemDiff.ChangedProperties, "marshalling should have resulted in non-nil value")

	changedNetworkInterfaces := res.ConfigurationItemDiff.getChangedNetworkInterfaces()
	assert.NotNil(t, changedNetworkInterfaces, "should have got back a non-nil map")
	assert.Equal(t, len(changedNetworkInterfaces), 2, "expected map size of 2")
	assert.NotNil(t, changedNetworkInterfaces[0], "should have got back a non-nil map")
	assert.Equal(t, "eni-fde9493f", changedNetworkInterfaces[0].PreviousValue.NetworkInterfaceID, "expected equality")
	assert.Nil(t, changedNetworkInterfaces[0].UpdatedValue, "expected nil")
	assert.Equal(t, "DELETE", changedNetworkInterfaces[0].ChangeType, "expected equality")

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
	_, err := Handle(context.Background(), Input{Message: string(marshalled)})
	assert.NotNil(t, err, "expected non-nil")
}

func TestTransformGoldenPath(t *testing.T) {
	// a real payload:
	const filename = "awsconfigpayload.json"

	str, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read file '%s': %s", filename, err)
	}

	res := awsConfigEvent{}
	json.Unmarshal(str, &res)

	changeTime, _ := time.Parse(time.RFC3339, "2017-01-09T22:50:14.328Z")

	expectedOutput := Output{
		AccountID:    "123456789012",
		ChangeTime:   changeTime, // from configurationItemCaptureTime // TODO: ok?
		Region:       "us-east-2",
		ResourceID:   "i-007d374c8912e3e90",
		ResourceType: "AWS::EC2::Instance",
		Tags:         map[string]string{"Name": "value"},
		Changes:      []Change{}}

	expectedChange0 := Change{}
	expectedChange0.Hostnames = []string{"ip-172-31-16-84.ec2.internal", "ec2-54-175-43-43.compute-1.amazonaws.com"}
	expectedChange0.PublicIPAddresses = []string{"54.175.43.43"}
	expectedChange0.PrivateIPAddresses = []string{"172.31.16.84"}
	expectedChange0.ChangeType = "DELETED"

	expectedOutput.Changes = append(expectedOutput.Changes, expectedChange0)

	expectedChange1 := Change{}
	expectedChange1.Hostnames = []string{"ip-172-31-16-84.ec2.internal", "ec2-54-175-43-43.compute-1.amazonaws.com"}
	expectedChange1.PublicIPAddresses = []string{"54.175.43.43"}
	expectedChange1.PrivateIPAddresses = []string{"172.31.16.84"}
	expectedChange1.ChangeType = "ADDED"

	expectedOutput.Changes = append(expectedOutput.Changes, expectedChange1)

	output, err := Handle(nil, Input{Message: string(str)})

	assert.Nil(t, err, "expected nil")
	assert.NotNil(t, output)

	assert.Equal(t, expectedOutput, output)

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
