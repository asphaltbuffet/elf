// Code generated by mockery v2.43.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MockConfigurationReader is an autogenerated mock type for the ConfigurationReader type
type MockConfigurationReader struct {
	mock.Mock
}

type MockConfigurationReader_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConfigurationReader) EXPECT() *MockConfigurationReader_Expecter {
	return &MockConfigurationReader_Expecter{mock: &_m.Mock}
}

// GetConfigFileUsed provides a mock function with given fields:
func (_m *MockConfigurationReader) GetConfigFileUsed() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetConfigFileUsed")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockConfigurationReader_GetConfigFileUsed_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConfigFileUsed'
type MockConfigurationReader_GetConfigFileUsed_Call struct {
	*mock.Call
}

// GetConfigFileUsed is a helper method to define mock.On call
func (_e *MockConfigurationReader_Expecter) GetConfigFileUsed() *MockConfigurationReader_GetConfigFileUsed_Call {
	return &MockConfigurationReader_GetConfigFileUsed_Call{Call: _e.mock.On("GetConfigFileUsed")}
}

func (_c *MockConfigurationReader_GetConfigFileUsed_Call) Run(run func()) *MockConfigurationReader_GetConfigFileUsed_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfigurationReader_GetConfigFileUsed_Call) Return(_a0 string) *MockConfigurationReader_GetConfigFileUsed_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigurationReader_GetConfigFileUsed_Call) RunAndReturn(run func() string) *MockConfigurationReader_GetConfigFileUsed_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConfigurationReader creates a new instance of MockConfigurationReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConfigurationReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConfigurationReader {
	mock := &MockConfigurationReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
