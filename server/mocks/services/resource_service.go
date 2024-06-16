// Code generated by MockGen. DO NOT EDIT.
// Source: resource_service.go

// Package services is a generated GoMock package.
package services

import (
	context "context"
	reflect "reflect"
    gomock "github.com/golang/mock/gomock"

	model "github.com/stsg/gophkeeper2/server/model"
	enum "github.com/stsg/gophkeeper2/pkg/model/enum"
)

// MockResourceService is a mock of ResourceService interface.
type MockResourceService struct {
	ctrl     *gomock.Controller
	recorder *MockResourceServiceMockRecorder
}

// MockResourceServiceMockRecorder is the mock recorder for MockResourceService.
type MockResourceServiceMockRecorder struct {
	mock *MockResourceService
}

// NewMockResourceService creates a new mock instance.
func NewMockResourceService(ctrl *gomock.Controller) *MockResourceService {
	mock := &MockResourceService{ctrl: ctrl}
	mock.recorder = &MockResourceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceService) EXPECT() *MockResourceServiceMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockResourceService) Delete(ctx context.Context, resId, userId int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, resId, userId)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockResourceServiceMockRecorder) Delete(ctx, resId, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockResourceService)(nil).Delete), ctx, resId, userId)
}

// Get mocks base method.
func (m *MockResourceService) Get(ctx context.Context, resId, userId int32) (*model.Resource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, resId, userId)
	ret0, _ := ret[0].(*model.Resource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockResourceServiceMockRecorder) Get(ctx, resId, userId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockResourceService)(nil).Get), ctx, resId, userId)
}

// GetDescriptions mocks base method.
func (m *MockResourceService) GetDescriptions(ctx context.Context, userId int32, resType enum.ResourceType) ([]*model.ResourceDescription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDescriptions", ctx, userId, resType)
	ret0, _ := ret[0].([]*model.ResourceDescription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDescriptions indicates an expected call of GetDescriptions.
func (mr *MockResourceServiceMockRecorder) GetDescriptions(ctx, userId, resType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDescriptions", reflect.TypeOf((*MockResourceService)(nil).GetDescriptions), ctx, userId, resType)
}

// GetFileDescription mocks base method.
func (m *MockResourceService) GetFileDescription(ctx context.Context, resource *model.Resource) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileDescription", ctx, resource)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileDescription indicates an expected call of GetFileDescription.
func (mr *MockResourceServiceMockRecorder) GetFileDescription(ctx, resource interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileDescription", reflect.TypeOf((*MockResourceService)(nil).GetFileDescription), ctx, resource)
}

// Save mocks base method.
func (m *MockResourceService) Save(ctx context.Context, res *model.Resource) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, res)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockResourceServiceMockRecorder) Save(ctx, res interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockResourceService)(nil).Save), ctx, res)
}

// SaveFileDescription mocks base method.
func (m *MockResourceService) SaveFileDescription(ctx context.Context, userId int32, meta, data []byte) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveFileDescription", ctx, userId, meta, data)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveFileDescription indicates an expected call of SaveFileDescription.
func (mr *MockResourceServiceMockRecorder) SaveFileDescription(ctx, userId, meta, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveFileDescription", reflect.TypeOf((*MockResourceService)(nil).SaveFileDescription), ctx, userId, meta, data)
}

// Update mocks base method.
func (m *MockResourceService) Update(ctx context.Context, res *model.Resource) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, res)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockResourceServiceMockRecorder) Update(ctx, res interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockResourceService)(nil).Update), ctx, res)
}
