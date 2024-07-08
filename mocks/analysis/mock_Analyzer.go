// Code generated by mockery v2.43.1. DO NOT EDIT.

package mocks

import (
	analysis "github.com/asphaltbuffet/elf/pkg/analysis"
	mock "github.com/stretchr/testify/mock"
)

// MockAnalyzer is an autogenerated mock type for the Analyzer type
type MockAnalyzer struct {
	mock.Mock
}

type MockAnalyzer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAnalyzer) EXPECT() *MockAnalyzer_Expecter {
	return &MockAnalyzer_Expecter{mock: &_m.Mock}
}

// Graph provides a mock function with given fields: _a0
func (_m *MockAnalyzer) Graph(_a0 analysis.GraphType) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Graph")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(analysis.GraphType) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAnalyzer_Graph_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Graph'
type MockAnalyzer_Graph_Call struct {
	*mock.Call
}

// Graph is a helper method to define mock.On call
//   - _a0 analysis.GraphType
func (_e *MockAnalyzer_Expecter) Graph(_a0 interface{}) *MockAnalyzer_Graph_Call {
	return &MockAnalyzer_Graph_Call{Call: _e.mock.On("Graph", _a0)}
}

func (_c *MockAnalyzer_Graph_Call) Run(run func(_a0 analysis.GraphType)) *MockAnalyzer_Graph_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(analysis.GraphType))
	})
	return _c
}

func (_c *MockAnalyzer_Graph_Call) Return(_a0 error) *MockAnalyzer_Graph_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAnalyzer_Graph_Call) RunAndReturn(run func(analysis.GraphType) error) *MockAnalyzer_Graph_Call {
	_c.Call.Return(run)
	return _c
}

// Stats provides a mock function with given fields:
func (_m *MockAnalyzer) Stats() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stats")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAnalyzer_Stats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stats'
type MockAnalyzer_Stats_Call struct {
	*mock.Call
}

// Stats is a helper method to define mock.On call
func (_e *MockAnalyzer_Expecter) Stats() *MockAnalyzer_Stats_Call {
	return &MockAnalyzer_Stats_Call{Call: _e.mock.On("Stats")}
}

func (_c *MockAnalyzer_Stats_Call) Run(run func()) *MockAnalyzer_Stats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAnalyzer_Stats_Call) Return(_a0 error) *MockAnalyzer_Stats_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAnalyzer_Stats_Call) RunAndReturn(run func() error) *MockAnalyzer_Stats_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAnalyzer creates a new instance of MockAnalyzer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAnalyzer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAnalyzer {
	mock := &MockAnalyzer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}