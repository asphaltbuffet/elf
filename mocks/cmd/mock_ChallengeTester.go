// Code generated by mockery v2.43.1. DO NOT EDIT.

package mocks

import (
	tasks "github.com/asphaltbuffet/elf/pkg/tasks"
	mock "github.com/stretchr/testify/mock"
)

// MockChallengeTester is an autogenerated mock type for the ChallengeTester type
type MockChallengeTester struct {
	mock.Mock
}

type MockChallengeTester_Expecter struct {
	mock *mock.Mock
}

func (_m *MockChallengeTester) EXPECT() *MockChallengeTester_Expecter {
	return &MockChallengeTester_Expecter{mock: &_m.Mock}
}

// String provides a mock function with given fields:
func (_m *MockChallengeTester) String() string {
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

// MockChallengeTester_String_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'String'
type MockChallengeTester_String_Call struct {
	*mock.Call
}

// String is a helper method to define mock.On call
func (_e *MockChallengeTester_Expecter) String() *MockChallengeTester_String_Call {
	return &MockChallengeTester_String_Call{Call: _e.mock.On("String")}
}

func (_c *MockChallengeTester_String_Call) Run(run func()) *MockChallengeTester_String_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockChallengeTester_String_Call) Return(_a0 string) *MockChallengeTester_String_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockChallengeTester_String_Call) RunAndReturn(run func() string) *MockChallengeTester_String_Call {
	_c.Call.Return(run)
	return _c
}

// Test provides a mock function with given fields:
func (_m *MockChallengeTester) Test() ([]tasks.Result, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Test")
	}

	var r0 []tasks.Result
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]tasks.Result, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []tasks.Result); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]tasks.Result)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockChallengeTester_Test_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Test'
type MockChallengeTester_Test_Call struct {
	*mock.Call
}

// Test is a helper method to define mock.On call
func (_e *MockChallengeTester_Expecter) Test() *MockChallengeTester_Test_Call {
	return &MockChallengeTester_Test_Call{Call: _e.mock.On("Test")}
}

func (_c *MockChallengeTester_Test_Call) Run(run func()) *MockChallengeTester_Test_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockChallengeTester_Test_Call) Return(_a0 []tasks.Result, _a1 error) *MockChallengeTester_Test_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockChallengeTester_Test_Call) RunAndReturn(run func() ([]tasks.Result, error)) *MockChallengeTester_Test_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockChallengeTester creates a new instance of MockChallengeTester. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockChallengeTester(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockChallengeTester {
	mock := &MockChallengeTester{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}