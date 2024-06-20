// Code generated by MockGen. DO NOT EDIT.
// Source: token_service.go

// Package services is a generated GoMock package.
package services

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockTokenService is a mock of TokenService interface.
type MockTokenService struct {
	ctrl     *gomock.Controller
	recorder *MockTokenServiceMockRecorder
}

// MockTokenServiceMockRecorder is the mock recorder for MockTokenService.
type MockTokenServiceMockRecorder struct {
	mock *MockTokenService
}

// NewMockTokenService creates a new mock instance.
func NewMockTokenService(ctrl *gomock.Controller) *MockTokenService {
	mock := &MockTokenService{ctrl: ctrl}
	mock.recorder = &MockTokenServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenService) EXPECT() *MockTokenServiceMockRecorder {
	return m.recorder
}

// ExtractUserId mocks base method.
func (m *MockTokenService) ExtractUserId(ctx context.Context) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtractUserId", ctx)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExtractUserId indicates an expected call of ExtractUserId.
func (mr *MockTokenServiceMockRecorder) ExtractUserId(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExtractUserId", reflect.TypeOf((*MockTokenService)(nil).ExtractUserId), ctx)
}

// Generate mocks base method.
func (m *MockTokenService) Generate(id int32, expireAt time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", id, expireAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockTokenServiceMockRecorder) Generate(id, expireAt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockTokenService)(nil).Generate), id, expireAt)
}
