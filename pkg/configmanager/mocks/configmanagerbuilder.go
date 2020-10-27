// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openshift/managed-upgrade-operator/pkg/configmanager (interfaces: ConfigManagerBuilder)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	configmanager "github.com/openshift/managed-upgrade-operator/pkg/configmanager"
	reflect "reflect"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockConfigManagerBuilder is a mock of ConfigManagerBuilder interface
type MockConfigManagerBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockConfigManagerBuilderMockRecorder
}

// MockConfigManagerBuilderMockRecorder is the mock recorder for MockConfigManagerBuilder
type MockConfigManagerBuilderMockRecorder struct {
	mock *MockConfigManagerBuilder
}

// NewMockConfigManagerBuilder creates a new mock instance
func NewMockConfigManagerBuilder(ctrl *gomock.Controller) *MockConfigManagerBuilder {
	mock := &MockConfigManagerBuilder{ctrl: ctrl}
	mock.recorder = &MockConfigManagerBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigManagerBuilder) EXPECT() *MockConfigManagerBuilderMockRecorder {
	return m.recorder
}

// New mocks base method
func (m *MockConfigManagerBuilder) New(arg0 client.Client, arg1 string) configmanager.ConfigManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", arg0, arg1)
	ret0, _ := ret[0].(configmanager.ConfigManager)
	return ret0
}

// New indicates an expected call of New
func (mr *MockConfigManagerBuilderMockRecorder) New(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockConfigManagerBuilder)(nil).New), arg0, arg1)
}
