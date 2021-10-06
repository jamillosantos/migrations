// Code generated by MockGen. DO NOT EDIT.
// Source: io/fs (interfaces: DirEntry)

// Package sql is a generated GoMock package.
package sql

import (
	fs "io/fs"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDirEntry is a mock of DirEntry interface.
type MockDirEntry struct {
	ctrl     *gomock.Controller
	recorder *MockDirEntryMockRecorder
}

// MockDirEntryMockRecorder is the mock recorder for MockDirEntry.
type MockDirEntryMockRecorder struct {
	mock *MockDirEntry
}

// NewMockDirEntry creates a new mock instance.
func NewMockDirEntry(ctrl *gomock.Controller) *MockDirEntry {
	mock := &MockDirEntry{ctrl: ctrl}
	mock.recorder = &MockDirEntryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDirEntry) EXPECT() *MockDirEntryMockRecorder {
	return m.recorder
}

// Info mocks base method.
func (m *MockDirEntry) Info() (fs.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(fs.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Info indicates an expected call of Info.
func (mr *MockDirEntryMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockDirEntry)(nil).Info))
}

// IsDir mocks base method.
func (m *MockDirEntry) IsDir() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDir")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDir indicates an expected call of IsDir.
func (mr *MockDirEntryMockRecorder) IsDir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDir", reflect.TypeOf((*MockDirEntry)(nil).IsDir))
}

// Name mocks base method.
func (m *MockDirEntry) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockDirEntryMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockDirEntry)(nil).Name))
}

// Type mocks base method.
func (m *MockDirEntry) Type() fs.FileMode {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Type")
	ret0, _ := ret[0].(fs.FileMode)
	return ret0
}

// Type indicates an expected call of Type.
func (mr *MockDirEntryMockRecorder) Type() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Type", reflect.TypeOf((*MockDirEntry)(nil).Type))
}
