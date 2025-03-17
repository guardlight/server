// Code generated by mockery v2.51.1. DO NOT EDIT.

package analysismanager

import (
	jobmanager "github.com/guardlight/server/internal/jobmanager"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// Mockjobber is an autogenerated mock type for the jobber type
type Mockjobber struct {
	mock.Mock
}

type Mockjobber_Expecter struct {
	mock *mock.Mock
}

func (_m *Mockjobber) EXPECT() *Mockjobber_Expecter {
	return &Mockjobber_Expecter{mock: &_m.Mock}
}

// CreateId provides a mock function with no fields
func (_m *Mockjobber) CreateId() uuid.UUID {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CreateId")
	}

	var r0 uuid.UUID
	if rf, ok := ret.Get(0).(func() uuid.UUID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	return r0
}

// Mockjobber_CreateId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateId'
type Mockjobber_CreateId_Call struct {
	*mock.Call
}

// CreateId is a helper method to define mock.On call
func (_e *Mockjobber_Expecter) CreateId() *Mockjobber_CreateId_Call {
	return &Mockjobber_CreateId_Call{Call: _e.mock.On("CreateId")}
}

func (_c *Mockjobber_CreateId_Call) Run(run func()) *Mockjobber_CreateId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Mockjobber_CreateId_Call) Return(_a0 uuid.UUID) *Mockjobber_CreateId_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Mockjobber_CreateId_Call) RunAndReturn(run func() uuid.UUID) *Mockjobber_CreateId_Call {
	_c.Call.Return(run)
	return _c
}

// EnqueueJob provides a mock function with given fields: id, jType, groupKey, data
func (_m *Mockjobber) EnqueueJob(id uuid.UUID, jType jobmanager.JobType, groupKey string, data interface{}) error {
	ret := _m.Called(id, jType, groupKey, data)

	if len(ret) == 0 {
		panic("no return value specified for EnqueueJob")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID, jobmanager.JobType, string, interface{}) error); ok {
		r0 = rf(id, jType, groupKey, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Mockjobber_EnqueueJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnqueueJob'
type Mockjobber_EnqueueJob_Call struct {
	*mock.Call
}

// EnqueueJob is a helper method to define mock.On call
//   - id uuid.UUID
//   - jType jobmanager.JobType
//   - groupKey string
//   - data interface{}
func (_e *Mockjobber_Expecter) EnqueueJob(id interface{}, jType interface{}, groupKey interface{}, data interface{}) *Mockjobber_EnqueueJob_Call {
	return &Mockjobber_EnqueueJob_Call{Call: _e.mock.On("EnqueueJob", id, jType, groupKey, data)}
}

func (_c *Mockjobber_EnqueueJob_Call) Run(run func(id uuid.UUID, jType jobmanager.JobType, groupKey string, data interface{})) *Mockjobber_EnqueueJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID), args[1].(jobmanager.JobType), args[2].(string), args[3].(interface{}))
	})
	return _c
}

func (_c *Mockjobber_EnqueueJob_Call) Return(_a0 error) *Mockjobber_EnqueueJob_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Mockjobber_EnqueueJob_Call) RunAndReturn(run func(uuid.UUID, jobmanager.JobType, string, interface{}) error) *Mockjobber_EnqueueJob_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateJobStatus provides a mock function with given fields: id, status, desc, retryCount
func (_m *Mockjobber) UpdateJobStatus(id uuid.UUID, status jobmanager.JobStatus, desc string, retryCount int) error {
	ret := _m.Called(id, status, desc, retryCount)

	if len(ret) == 0 {
		panic("no return value specified for UpdateJobStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uuid.UUID, jobmanager.JobStatus, string, int) error); ok {
		r0 = rf(id, status, desc, retryCount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Mockjobber_UpdateJobStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateJobStatus'
type Mockjobber_UpdateJobStatus_Call struct {
	*mock.Call
}

// UpdateJobStatus is a helper method to define mock.On call
//   - id uuid.UUID
//   - status jobmanager.JobStatus
//   - desc string
//   - retryCount int
func (_e *Mockjobber_Expecter) UpdateJobStatus(id interface{}, status interface{}, desc interface{}, retryCount interface{}) *Mockjobber_UpdateJobStatus_Call {
	return &Mockjobber_UpdateJobStatus_Call{Call: _e.mock.On("UpdateJobStatus", id, status, desc, retryCount)}
}

func (_c *Mockjobber_UpdateJobStatus_Call) Run(run func(id uuid.UUID, status jobmanager.JobStatus, desc string, retryCount int)) *Mockjobber_UpdateJobStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID), args[1].(jobmanager.JobStatus), args[2].(string), args[3].(int))
	})
	return _c
}

func (_c *Mockjobber_UpdateJobStatus_Call) Return(_a0 error) *Mockjobber_UpdateJobStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Mockjobber_UpdateJobStatus_Call) RunAndReturn(run func(uuid.UUID, jobmanager.JobStatus, string, int) error) *Mockjobber_UpdateJobStatus_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockjobber creates a new instance of Mockjobber. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockjobber(t interface {
	mock.TestingT
	Cleanup(func())
}) *Mockjobber {
	mock := &Mockjobber{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
