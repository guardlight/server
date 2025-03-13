// Code generated by mockery v2.51.1. DO NOT EDIT.

package jobmanager

import mock "github.com/stretchr/testify/mock"

// MockjobSaver is an autogenerated mock type for the jobSaver type
type MockjobSaver struct {
	mock.Mock
}

type MockjobSaver_Expecter struct {
	mock *mock.Mock
}

func (_m *MockjobSaver) EXPECT() *MockjobSaver_Expecter {
	return &MockjobSaver_Expecter{mock: &_m.Mock}
}

// saveJob provides a mock function with given fields: j
func (_m *MockjobSaver) saveJob(j *Job) error {
	ret := _m.Called(j)

	if len(ret) == 0 {
		panic("no return value specified for saveJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*Job) error); ok {
		r0 = rf(j)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockjobSaver_saveJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'saveJob'
type MockjobSaver_saveJob_Call struct {
	*mock.Call
}

// saveJob is a helper method to define mock.On call
//   - j *Job
func (_e *MockjobSaver_Expecter) saveJob(j interface{}) *MockjobSaver_saveJob_Call {
	return &MockjobSaver_saveJob_Call{Call: _e.mock.On("saveJob", j)}
}

func (_c *MockjobSaver_saveJob_Call) Run(run func(j *Job)) *MockjobSaver_saveJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Job))
	})
	return _c
}

func (_c *MockjobSaver_saveJob_Call) Return(_a0 error) *MockjobSaver_saveJob_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockjobSaver_saveJob_Call) RunAndReturn(run func(*Job) error) *MockjobSaver_saveJob_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockjobSaver creates a new instance of MockjobSaver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockjobSaver(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockjobSaver {
	mock := &MockjobSaver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
