// Code generated by mockery v2.51.1. DO NOT EDIT.

package theme

import (
	uuid "github.com/google/uuid"
	mock "github.com/stretchr/testify/mock"
)

// MockthemeStore is an autogenerated mock type for the themeStore type
type MockthemeStore struct {
	mock.Mock
}

type MockthemeStore_Expecter struct {
	mock *mock.Mock
}

func (_m *MockthemeStore) EXPECT() *MockthemeStore_Expecter {
	return &MockthemeStore_Expecter{mock: &_m.Mock}
}

// getAllThemesByUserId provides a mock function with given fields: id
func (_m *MockthemeStore) getAllThemesByUserId(id uuid.UUID) ([]Theme, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for getAllThemesByUserId")
	}

	var r0 []Theme
	var r1 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) ([]Theme, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(uuid.UUID) []Theme); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Theme)
		}
	}

	if rf, ok := ret.Get(1).(func(uuid.UUID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockthemeStore_getAllThemesByUserId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'getAllThemesByUserId'
type MockthemeStore_getAllThemesByUserId_Call struct {
	*mock.Call
}

// getAllThemesByUserId is a helper method to define mock.On call
//   - id uuid.UUID
func (_e *MockthemeStore_Expecter) getAllThemesByUserId(id interface{}) *MockthemeStore_getAllThemesByUserId_Call {
	return &MockthemeStore_getAllThemesByUserId_Call{Call: _e.mock.On("getAllThemesByUserId", id)}
}

func (_c *MockthemeStore_getAllThemesByUserId_Call) Run(run func(id uuid.UUID)) *MockthemeStore_getAllThemesByUserId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID))
	})
	return _c
}

func (_c *MockthemeStore_getAllThemesByUserId_Call) Return(_a0 []Theme, _a1 error) *MockthemeStore_getAllThemesByUserId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockthemeStore_getAllThemesByUserId_Call) RunAndReturn(run func(uuid.UUID) ([]Theme, error)) *MockthemeStore_getAllThemesByUserId_Call {
	_c.Call.Return(run)
	return _c
}

// updateTheme provides a mock function with given fields: t, uid
func (_m *MockthemeStore) updateTheme(t *Theme, uid uuid.UUID) error {
	ret := _m.Called(t, uid)

	if len(ret) == 0 {
		panic("no return value specified for updateTheme")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*Theme, uuid.UUID) error); ok {
		r0 = rf(t, uid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockthemeStore_updateTheme_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'updateTheme'
type MockthemeStore_updateTheme_Call struct {
	*mock.Call
}

// updateTheme is a helper method to define mock.On call
//   - t *Theme
//   - uid uuid.UUID
func (_e *MockthemeStore_Expecter) updateTheme(t interface{}, uid interface{}) *MockthemeStore_updateTheme_Call {
	return &MockthemeStore_updateTheme_Call{Call: _e.mock.On("updateTheme", t, uid)}
}

func (_c *MockthemeStore_updateTheme_Call) Run(run func(t *Theme, uid uuid.UUID)) *MockthemeStore_updateTheme_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*Theme), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *MockthemeStore_updateTheme_Call) Return(_a0 error) *MockthemeStore_updateTheme_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockthemeStore_updateTheme_Call) RunAndReturn(run func(*Theme, uuid.UUID) error) *MockthemeStore_updateTheme_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockthemeStore creates a new instance of MockthemeStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockthemeStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockthemeStore {
	mock := &MockthemeStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
