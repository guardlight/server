// Code generated by mockery v2.51.1. DO NOT EDIT.

package jobmanager

import (
	uuid "github.com/google/uuid"
	mock "github.com/stretchr/testify/mock"
)

// MockjobStore is an autogenerated mock type for the jobStore type
type MockjobStore struct {
	mock.Mock
}

type MockjobStore_Expecter struct {
	mock *mock.Mock
}

func (_m *MockjobStore) EXPECT() *MockjobStore_Expecter {
	return &MockjobStore_Expecter{mock: &_m.Mock}
}

// getNotFinishedJobs provides a mock function with no fields
func (_m *MockjobStore) getNotFinishedJobs() ([]Job, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for getNotFinishedJobs")
	}

	var r0 []Job
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]Job, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []Job); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Job)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockjobStore_getNotFinishedJobs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'getNotFinishedJobs'
type MockjobStore_getNotFinishedJobs_Call struct {
	*mock.Call
}

// getNotFinishedJobs is a helper method to define mock.On call
func (_e *MockjobStore_Expecter) getNotFinishedJobs() *MockjobStore_getNotFinishedJobs_Call {
	return &MockjobStore_getNotFinishedJobs_Call{Call: _e.mock.On("getNotFinishedJobs")}
}

func (_c *MockjobStore_getNotFinishedJobs_Call) Run(run func()) *MockjobStore_getNotFinishedJobs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockjobStore_getNotFinishedJobs_Call) Return(_a0 []Job, _a1 error) *MockjobStore_getNotFinishedJobs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockjobStore_getNotFinishedJobs_Call) RunAndReturn(run func() ([]Job, error)) *MockjobStore_getNotFinishedJobs_Call {
	_c.Call.Return(run)
	return _c
}

// saveJob provides a mock function with given fields: j
func (_m *MockjobStore) saveJob(j *Job) error {
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

// MockjobStore_saveJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'saveJob'
type MockjobStore_saveJob_Call struct {
	*mock.Call
}

// saveJob is a helper method to define mock.On call
//   - j *Job
func (_e *MockjobStore_Expecter) saveJob(j interface{}) *MockjobStore_saveJob_Call {
	return &MockjobStore_saveJob_Call{Call: _e.mock.On("saveJob", j)}
}

func (_c *MockjobStore_saveJob_Call) Run(run func(j *Job)) *MockjobStore_saveJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Job))
	})
	return _c
}

func (_c *MockjobStore_saveJob_Call) Return(_a0 error) *MockjobStore_saveJob_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockjobStore_saveJob_Call) RunAndReturn(run func(*Job) error) *MockjobStore_saveJob_Call {
	_c.Call.Return(run)
	return _c
}

// updateJobStatus provides a mock function with given fields: id, s, sd, rc
func (_m *MockjobStore) updateJobStatus(id uuid.UUID, s JobStatus, sd string, rc int) error {
	ret := _m.Called(id, s, sd, rc)

	if len(ret) == 0 {
		panic("no return value specified for updateJobStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID, JobStatus, string, int) error); ok {
		r0 = rf(id, s, sd, rc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockjobStore_updateJobStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'updateJobStatus'
type MockjobStore_updateJobStatus_Call struct {
	*mock.Call
}

// updateJobStatus is a helper method to define mock.On call
//   - id uuid.UUID
//   - s JobStatus
//   - sd string
//   - rc int
func (_e *MockjobStore_Expecter) updateJobStatus(id interface{}, s interface{}, sd interface{}, rc interface{}) *MockjobStore_updateJobStatus_Call {
	return &MockjobStore_updateJobStatus_Call{Call: _e.mock.On("updateJobStatus", id, s, sd, rc)}
}

func (_c *MockjobStore_updateJobStatus_Call) Run(run func(id uuid.UUID, s JobStatus, sd string, rc int)) *MockjobStore_updateJobStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID), args[1].(JobStatus), args[2].(string), args[3].(int))
	})
	return _c
}

func (_c *MockjobStore_updateJobStatus_Call) Return(_a0 error) *MockjobStore_updateJobStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockjobStore_updateJobStatus_Call) RunAndReturn(run func(uuid.UUID, JobStatus, string, int) error) *MockjobStore_updateJobStatus_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockjobStore creates a new instance of MockjobStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockjobStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockjobStore {
	mock := &MockjobStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
