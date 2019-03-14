package v1

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
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

	res := AWSConfigEvent{}
	json.Unmarshal(str, &res)

	// spot check a few things to confirm both static marshalling ^^^ and dynamic (via get* functions) all work

	assert.NotNil(t, res.ConfigurationItemDiff.ChangedProperties, "marshalling should have resulted in non-nil value")

	changedNetworkInterfaces := res.ConfigurationItemDiff.getChangedNetworkInterfaces()
	assert.NotNil(t, changedNetworkInterfaces, "should have got back a non-nil map")
	assert.Equal(t, len(changedNetworkInterfaces), 2, "expected map size of 2")
	assert.NotNil(t, changedNetworkInterfaces["Configuration.NetworkInterfaces.0"], "should have got back a non-nil map")
	assert.Equal(t, "eni-fde9493f", changedNetworkInterfaces["Configuration.NetworkInterfaces.0"].PreviousValue.NetworkInterfaceID, "expected equality")
	assert.Nil(t, changedNetworkInterfaces["Configuration.NetworkInterfaces.0"].UpdatedValue, "expected nil")
	assert.Equal(t, "DELETE", changedNetworkInterfaces["Configuration.NetworkInterfaces.0"].ChangeType, "expected equality")

	securityGroups := res.ConfigurationItemDiff.getConfigurationSecurityGroups()
	assert.NotNil(t, securityGroups, "should have got back a non-nil map")
	assert.Equal(t, len(securityGroups), 2, "expected map size of 2")
	assert.Nil(t, securityGroups["Configuration.SecurityGroups.1"].PreviousValue, "expected nil")
	assert.NotNil(t, securityGroups["Configuration.SecurityGroups.1"].UpdatedValue, "expected non-nil")
	assert.Equal(t, "example-security-group-2", securityGroups["Configuration.SecurityGroups.1"].UpdatedValue.GroupName, "expected equality")

	relationships := res.ConfigurationItemDiff.getRelationships()
	assert.NotNil(t, relationships, "should have got back a non-nil map")
	assert.Equal(t, len(relationships), 2, "expected map size of 2")
	assert.Nil(t, relationships["Relationships.0"].UpdatedValue, "expected nil")
	assert.NotNil(t, relationships["Relationships.0"].PreviousValue, "expected non-nil")
	assert.Equal(t, "sg-c8b141b4", relationships["Relationships.0"].PreviousValue.ResourceID, "expected equality")

	assert.Empty(t, res.ConfigurationItem.Configuration.StateTransitionReason, "expected empty string")
	assert.Nil(t, res.ConfigurationItem.Configuration.KernelID, "expected nil")

	marshalled, _ := json.MarshalIndent(res, "", "    ")
	// if you want to see it:
	//fmt.Println(string(marshalled))

	// don't worry about this
	modified := strings.ReplaceAll(string(str), "changedProperties", "ChangedProperties")
	jsonEquals, _ := jsonBytesEqual([]byte(modified), marshalled)

	assert.True(t, jsonEquals, "expect exact JSON equality before marshall/unmarshal and after.  "+
		"If they're not, it's probably because your JSON key does not match the field name "+
		"or you did not use a pointer where you should have")

}

func TestTransform(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expectedOutput := Output{}

	mockObj := NewMockReporter(mockCtrl)
	mockObj.EXPECT().Report(gomock.Any(), gomock.Eq(expectedOutput)).Return(nil)

	handler := AWSConfigChangeEventHandler{
		Reporter: mockObj}

	err := handler.Handle(nil, AWSConfigEvent{})
	assert.Nil(t, err, "expected nil")
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
