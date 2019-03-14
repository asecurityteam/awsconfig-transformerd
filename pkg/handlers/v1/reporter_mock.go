// Code generated by MockGen. DO NOT EDIT.
// Source: reporter.go

// Package v1 is a generated GoMock package.
package v1

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockReporter is a mock of Reporter interface
type MockReporter struct {
	ctrl     *gomock.Controller
	recorder *MockReporterMockRecorder
}

// MockReporterMockRecorder is the mock recorder for MockReporter
type MockReporterMockRecorder struct {
	mock *MockReporter
}

// NewMockReporter creates a new mock instance
func NewMockReporter(ctrl *gomock.Controller) *MockReporter {
	mock := &MockReporter{ctrl: ctrl}
	mock.recorder = &MockReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReporter) EXPECT() *MockReporterMockRecorder {
	return m.recorder
}

// Report mocks base method
func (m *MockReporter) Report(ctx context.Context, payload Output) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Report", ctx, payload)
	ret0, _ := ret[0].(error)
	return ret0
}

// Report indicates an expected call of Report
func (mr *MockReporterMockRecorder) Report(ctx, payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Report", reflect.TypeOf((*MockReporter)(nil).Report), ctx, payload)
}
