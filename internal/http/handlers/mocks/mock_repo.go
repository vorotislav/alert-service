// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vorotislav/alert-service/internal/http/handlers (interfaces: Repository)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/vorotislav/alert-service/internal/model"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// AllMetrics mocks base method.
func (m *MockRepository) AllMetrics(arg0 context.Context) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllMetrics", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllMetrics indicates an expected call of AllMetrics.
func (mr *MockRepositoryMockRecorder) AllMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllMetrics", reflect.TypeOf((*MockRepository)(nil).AllMetrics), arg0)
}

// GetCounterValue mocks base method.
func (m *MockRepository) GetCounterValue(arg0 context.Context, arg1 string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounterValue", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCounterValue indicates an expected call of GetCounterValue.
func (mr *MockRepositoryMockRecorder) GetCounterValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounterValue", reflect.TypeOf((*MockRepository)(nil).GetCounterValue), arg0, arg1)
}

// GetGaugeValue mocks base method.
func (m *MockRepository) GetGaugeValue(arg0 context.Context, arg1 string) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGaugeValue", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGaugeValue indicates an expected call of GetGaugeValue.
func (mr *MockRepositoryMockRecorder) GetGaugeValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGaugeValue", reflect.TypeOf((*MockRepository)(nil).GetGaugeValue), arg0, arg1)
}

// Ping mocks base method.
func (m *MockRepository) Ping(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockRepositoryMockRecorder) Ping(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockRepository)(nil).Ping), arg0)
}

// UpdateMetric mocks base method.
func (m *MockRepository) UpdateMetric(arg0 context.Context, arg1 model.Metrics) (model.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMetric", arg0, arg1)
	ret0, _ := ret[0].(model.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateMetric indicates an expected call of UpdateMetric.
func (mr *MockRepositoryMockRecorder) UpdateMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMetric", reflect.TypeOf((*MockRepository)(nil).UpdateMetric), arg0, arg1)
}

// UpdateMetrics mocks base method.
func (m *MockRepository) UpdateMetrics(arg0 context.Context, arg1 []model.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMetrics", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateMetrics indicates an expected call of UpdateMetrics.
func (mr *MockRepositoryMockRecorder) UpdateMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMetrics", reflect.TypeOf((*MockRepository)(nil).UpdateMetrics), arg0, arg1)
}