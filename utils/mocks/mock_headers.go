// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ARM-software/golang-utils/utils/http/headers (interfaces: IHTTPHeaders)
//
// Generated by this command:
//
//	mockgen -destination=../../mocks/mock_headers.go -package=mocks github.com/ARM-software/golang-utils/utils/http/headers IHTTPHeaders
//

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	reflect "reflect"

	headers "github.com/ARM-software/golang-utils/utils/http/headers"
	gomock "go.uber.org/mock/gomock"
)

// MockIHTTPHeaders is a mock of IHTTPHeaders interface.
type MockIHTTPHeaders struct {
	ctrl     *gomock.Controller
	recorder *MockIHTTPHeadersMockRecorder
	isgomock struct{}
}

// MockIHTTPHeadersMockRecorder is the mock recorder for MockIHTTPHeaders.
type MockIHTTPHeadersMockRecorder struct {
	mock *MockIHTTPHeaders
}

// NewMockIHTTPHeaders creates a new mock instance.
func NewMockIHTTPHeaders(ctrl *gomock.Controller) *MockIHTTPHeaders {
	mock := &MockIHTTPHeaders{ctrl: ctrl}
	mock.recorder = &MockIHTTPHeadersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIHTTPHeaders) EXPECT() *MockIHTTPHeadersMockRecorder {
	return m.recorder
}

// Append mocks base method.
func (m *MockIHTTPHeaders) Append(h *headers.Header) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Append", h)
}

// Append indicates an expected call of Append.
func (mr *MockIHTTPHeadersMockRecorder) Append(h any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Append", reflect.TypeOf((*MockIHTTPHeaders)(nil).Append), h)
}

// AppendHeader mocks base method.
func (m *MockIHTTPHeaders) AppendHeader(key, value string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AppendHeader", key, value)
}

// AppendHeader indicates an expected call of AppendHeader.
func (mr *MockIHTTPHeadersMockRecorder) AppendHeader(key, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendHeader", reflect.TypeOf((*MockIHTTPHeaders)(nil).AppendHeader), key, value)
}

// AppendToResponse mocks base method.
func (m *MockIHTTPHeaders) AppendToResponse(w http.ResponseWriter) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AppendToResponse", w)
}

// AppendToResponse indicates an expected call of AppendToResponse.
func (mr *MockIHTTPHeadersMockRecorder) AppendToResponse(w any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendToResponse", reflect.TypeOf((*MockIHTTPHeaders)(nil).AppendToResponse), w)
}

// Empty mocks base method.
func (m *MockIHTTPHeaders) Empty() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Empty")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Empty indicates an expected call of Empty.
func (mr *MockIHTTPHeadersMockRecorder) Empty() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Empty", reflect.TypeOf((*MockIHTTPHeaders)(nil).Empty))
}

// Has mocks base method.
func (m *MockIHTTPHeaders) Has(h *headers.Header) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", h)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Has indicates an expected call of Has.
func (mr *MockIHTTPHeadersMockRecorder) Has(h any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockIHTTPHeaders)(nil).Has), h)
}

// HasHeader mocks base method.
func (m *MockIHTTPHeaders) HasHeader(key string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasHeader", key)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasHeader indicates an expected call of HasHeader.
func (mr *MockIHTTPHeadersMockRecorder) HasHeader(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasHeader", reflect.TypeOf((*MockIHTTPHeaders)(nil).HasHeader), key)
}
