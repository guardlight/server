// Code generated by mockery v2.51.1. DO NOT EDIT.

package analysismanager

import (
	theme "github.com/guardlight/server/internal/theme"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockthemeService is an autogenerated mock type for the themeService type
type MockthemeService struct {
	mock.Mock
}

type MockthemeService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockthemeService) EXPECT() *MockthemeService_Expecter {
	return &MockthemeService_Expecter{mock: &_m.Mock}
}

// GetAllThemesByUserId provides a mock function with given fields: id
func (_m *MockthemeService) GetAllThemesByUserId(id uuid.UUID) ([]theme.ThemeDto, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetAllThemesByUserId")
	}

	var r0 []theme.ThemeDto
	var r1 error
	if rf, ok := ret.Get(0).(func(uuid.UUID) ([]theme.ThemeDto, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(uuid.UUID) []theme.ThemeDto); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]theme.ThemeDto)
		}
	}

	if rf, ok := ret.Get(1).(func(uuid.UUID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockthemeService_GetAllThemesByUserId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllThemesByUserId'
type MockthemeService_GetAllThemesByUserId_Call struct {
	*mock.Call
}

// GetAllThemesByUserId is a helper method to define mock.On call
//   - id uuid.UUID
func (_e *MockthemeService_Expecter) GetAllThemesByUserId(id interface{}) *MockthemeService_GetAllThemesByUserId_Call {
	return &MockthemeService_GetAllThemesByUserId_Call{Call: _e.mock.On("GetAllThemesByUserId", id)}
}

func (_c *MockthemeService_GetAllThemesByUserId_Call) Run(run func(id uuid.UUID)) *MockthemeService_GetAllThemesByUserId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uuid.UUID))
	})
	return _c
}

func (_c *MockthemeService_GetAllThemesByUserId_Call) Return(_a0 []theme.ThemeDto, _a1 error) *MockthemeService_GetAllThemesByUserId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockthemeService_GetAllThemesByUserId_Call) RunAndReturn(run func(uuid.UUID) ([]theme.ThemeDto, error)) *MockthemeService_GetAllThemesByUserId_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockthemeService creates a new instance of MockthemeService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockthemeService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockthemeService {
	mock := &MockthemeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
