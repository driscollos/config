// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/driscollos/config/internal/sourcer/terminal-reader (interfaces: TerminalReader)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTerminalReader is a mock of TerminalReader interface.
type MockTerminalReader struct {
	ctrl     *gomock.Controller
	recorder *MockTerminalReaderMockRecorder
}

// MockTerminalReaderMockRecorder is the mock recorder for MockTerminalReader.
type MockTerminalReaderMockRecorder struct {
	mock *MockTerminalReader
}

// NewMockTerminalReader creates a new mock instance.
func NewMockTerminalReader(ctrl *gomock.Controller) *MockTerminalReader {
	mock := &MockTerminalReader{ctrl: ctrl}
	mock.recorder = &MockTerminalReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTerminalReader) EXPECT() *MockTerminalReaderMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockTerminalReader) Get(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockTerminalReaderMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockTerminalReader)(nil).Get), arg0)
}
