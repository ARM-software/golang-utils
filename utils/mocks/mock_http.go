// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ARM-software/golang-utils/utils/http (interfaces: IClient,IRetryWaitPolicy)

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	url "net/url"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockIClient is a mock of IClient interface.
type MockIClient struct {
	ctrl     *gomock.Controller
	recorder *MockIClientMockRecorder
}

// MockIClientMockRecorder is the mock recorder for MockIClient.
type MockIClientMockRecorder struct {
	mock *MockIClient
}

// NewMockIClient creates a new mock instance.
func NewMockIClient(ctrl *gomock.Controller) *MockIClient {
	mock := &MockIClient{ctrl: ctrl}
	mock.recorder = &MockIClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIClient) EXPECT() *MockIClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIClient)(nil).Close))
}

// Delete mocks base method.
func (m *MockIClient) Delete(arg0 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockIClientMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockIClient)(nil).Delete), arg0)
}

// Do mocks base method.
func (m *MockIClient) Do(arg0 *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockIClientMockRecorder) Do(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockIClient)(nil).Do), arg0)
}

// Get mocks base method.
func (m *MockIClient) Get(arg0 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockIClientMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockIClient)(nil).Get), arg0)
}

// Head mocks base method.
func (m *MockIClient) Head(arg0 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Head", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Head indicates an expected call of Head.
func (mr *MockIClientMockRecorder) Head(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Head", reflect.TypeOf((*MockIClient)(nil).Head), arg0)
}

// Options mocks base method.
func (m *MockIClient) Options(arg0 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Options", arg0)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Options indicates an expected call of Options.
func (mr *MockIClientMockRecorder) Options(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Options", reflect.TypeOf((*MockIClient)(nil).Options), arg0)
}

// Post mocks base method.
func (m *MockIClient) Post(arg0, arg1 string, arg2 interface{}) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Post", arg0, arg1, arg2)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Post indicates an expected call of Post.
func (mr *MockIClientMockRecorder) Post(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Post", reflect.TypeOf((*MockIClient)(nil).Post), arg0, arg1, arg2)
}

// PostForm mocks base method.
func (m *MockIClient) PostForm(arg0 string, arg1 url.Values) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostForm", arg0, arg1)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostForm indicates an expected call of PostForm.
func (mr *MockIClientMockRecorder) PostForm(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostForm", reflect.TypeOf((*MockIClient)(nil).PostForm), arg0, arg1)
}

// Put mocks base method.
func (m *MockIClient) Put(arg0 string, arg1 interface{}) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", arg0, arg1)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Put indicates an expected call of Put.
func (mr *MockIClientMockRecorder) Put(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockIClient)(nil).Put), arg0, arg1)
}

// StandardClient mocks base method.
func (m *MockIClient) StandardClient() *http.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StandardClient")
	ret0, _ := ret[0].(*http.Client)
	return ret0
}

// StandardClient indicates an expected call of StandardClient.
func (mr *MockIClientMockRecorder) StandardClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StandardClient", reflect.TypeOf((*MockIClient)(nil).StandardClient))
}

// MockIRetryWaitPolicy is a mock of IRetryWaitPolicy interface.
type MockIRetryWaitPolicy struct {
	ctrl     *gomock.Controller
	recorder *MockIRetryWaitPolicyMockRecorder
}

// MockIRetryWaitPolicyMockRecorder is the mock recorder for MockIRetryWaitPolicy.
type MockIRetryWaitPolicyMockRecorder struct {
	mock *MockIRetryWaitPolicy
}

// NewMockIRetryWaitPolicy creates a new mock instance.
func NewMockIRetryWaitPolicy(ctrl *gomock.Controller) *MockIRetryWaitPolicy {
	mock := &MockIRetryWaitPolicy{ctrl: ctrl}
	mock.recorder = &MockIRetryWaitPolicyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIRetryWaitPolicy) EXPECT() *MockIRetryWaitPolicyMockRecorder {
	return m.recorder
}

// Apply mocks base method.
func (m *MockIRetryWaitPolicy) Apply(arg0, arg1 time.Duration, arg2 int, arg3 *http.Response) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Apply", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// Apply indicates an expected call of Apply.
func (mr *MockIRetryWaitPolicyMockRecorder) Apply(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Apply", reflect.TypeOf((*MockIRetryWaitPolicy)(nil).Apply), arg0, arg1, arg2, arg3)
}
