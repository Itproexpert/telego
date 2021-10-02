// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mymmrac/go-telegram-bot-api/api (interfaces: Caller)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	api "github.com/mymmrac/go-telegram-bot-api/api"
)

// MockCaller is a mock of Caller interface.
type MockCaller struct {
	ctrl     *gomock.Controller
	recorder *MockCallerMockRecorder
}

// MockCallerMockRecorder is the mock recorder for MockCaller.
type MockCallerMockRecorder struct {
	mock *MockCaller
}

// NewMockCaller creates a new mock instance.
func NewMockCaller(ctrl *gomock.Controller) *MockCaller {
	mock := &MockCaller{ctrl: ctrl}
	mock.recorder = &MockCallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCaller) EXPECT() *MockCallerMockRecorder {
	return m.recorder
}

// Call mocks base method.
func (m *MockCaller) Call(arg0 string, arg1 *api.RequestData) (*api.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", arg0, arg1)
	ret0, _ := ret[0].(*api.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Call indicates an expected call of Call.
func (mr *MockCallerMockRecorder) Call(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockCaller)(nil).Call), arg0, arg1)
}
