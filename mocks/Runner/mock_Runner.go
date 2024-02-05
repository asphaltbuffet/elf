// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import (
	runners "github.com/asphaltbuffet/elf/pkg/runners"
	mock "github.com/stretchr/testify/mock"
)

// MockRunner is an autogenerated mock type for the Runner type
type MockRunner struct {
	mock.Mock
}

type MockRunner_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRunner) EXPECT() *MockRunner_Expecter {
	return &MockRunner_Expecter{mock: &_m.Mock}
}

// Cleanup provides a mock function with given fields:
func (_m *MockRunner) Cleanup() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Cleanup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRunner_Cleanup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cleanup'
type MockRunner_Cleanup_Call struct {
	*mock.Call
}

// Cleanup is a helper method to define mock.On call
func (_e *MockRunner_Expecter) Cleanup() *MockRunner_Cleanup_Call {
	return &MockRunner_Cleanup_Call{Call: _e.mock.On("Cleanup")}
}

func (_c *MockRunner_Cleanup_Call) Run(run func()) *MockRunner_Cleanup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRunner_Cleanup_Call) Return(_a0 error) *MockRunner_Cleanup_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRunner_Cleanup_Call) RunAndReturn(run func() error) *MockRunner_Cleanup_Call {
	_c.Call.Return(run)
	return _c
}

// Run provides a mock function with given fields: task
func (_m *MockRunner) Run(task *runners.Task) (*runners.Result, error) {
	ret := _m.Called(task)

	if len(ret) == 0 {
		panic("no return value specified for Run")
	}

	var r0 *runners.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(*runners.Task) (*runners.Result, error)); ok {
		return rf(task)
	}
	if rf, ok := ret.Get(0).(func(*runners.Task) *runners.Result); ok {
		r0 = rf(task)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*runners.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(*runners.Task) error); ok {
		r1 = rf(task)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRunner_Run_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Run'
type MockRunner_Run_Call struct {
	*mock.Call
}

// Run is a helper method to define mock.On call
//   - task *runners.Task
func (_e *MockRunner_Expecter) Run(task interface{}) *MockRunner_Run_Call {
	return &MockRunner_Run_Call{Call: _e.mock.On("Run", task)}
}

func (_c *MockRunner_Run_Call) Run(run func(task *runners.Task)) *MockRunner_Run_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*runners.Task))
	})
	return _c
}

func (_c *MockRunner_Run_Call) Return(_a0 *runners.Result, _a1 error) *MockRunner_Run_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRunner_Run_Call) RunAndReturn(run func(*runners.Task) (*runners.Result, error)) *MockRunner_Run_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields:
func (_m *MockRunner) Start() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRunner_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockRunner_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *MockRunner_Expecter) Start() *MockRunner_Start_Call {
	return &MockRunner_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *MockRunner_Start_Call) Run(run func()) *MockRunner_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRunner_Start_Call) Return(_a0 error) *MockRunner_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRunner_Start_Call) RunAndReturn(run func() error) *MockRunner_Start_Call {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with given fields:
func (_m *MockRunner) Stop() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRunner_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type MockRunner_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *MockRunner_Expecter) Stop() *MockRunner_Stop_Call {
	return &MockRunner_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *MockRunner_Stop_Call) Run(run func()) *MockRunner_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRunner_Stop_Call) Return(_a0 error) *MockRunner_Stop_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRunner_Stop_Call) RunAndReturn(run func() error) *MockRunner_Stop_Call {
	_c.Call.Return(run)
	return _c
}

// String provides a mock function with given fields:
func (_m *MockRunner) String() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for String")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockRunner_String_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'String'
type MockRunner_String_Call struct {
	*mock.Call
}

// String is a helper method to define mock.On call
func (_e *MockRunner_Expecter) String() *MockRunner_String_Call {
	return &MockRunner_String_Call{Call: _e.mock.On("String")}
}

func (_c *MockRunner_String_Call) Run(run func()) *MockRunner_String_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRunner_String_Call) Return(_a0 string) *MockRunner_String_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRunner_String_Call) RunAndReturn(run func() string) *MockRunner_String_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRunner creates a new instance of MockRunner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRunner(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRunner {
	mock := &MockRunner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}