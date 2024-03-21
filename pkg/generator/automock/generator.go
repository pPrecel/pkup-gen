// Code generated by mockery v2.33.3. DO NOT EDIT.

package automock

import (
	generator "github.com/pPrecel/PKUP/pkg/generator"
	config "github.com/pPrecel/PKUP/pkg/generator/config"

	mock "github.com/stretchr/testify/mock"
)

// Generator is an autogenerated mock type for the Generator type
type Generator struct {
	mock.Mock
}

// ForArgs provides a mock function with given fields: _a0
func (_m *Generator) ForArgs(_a0 *generator.GeneratorArgs) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*generator.GeneratorArgs) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ForConfig provides a mock function with given fields: _a0, _a1
func (_m *Generator) ForConfig(_a0 *config.Config, _a1 generator.ComposeOpts) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(*config.Config, generator.ComposeOpts) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewGenerator creates a new instance of Generator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGenerator(t interface {
	mock.TestingT
	Cleanup(func())
}) *Generator {
	mock := &Generator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
