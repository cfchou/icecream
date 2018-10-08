// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// APIKeyBackend is an autogenerated mock type for the APIKeyBackend type
type APIKeyBackend struct {
	mock.Mock
}

// Authenticate provides a mock function with given fields: apiKey
func (_m *APIKeyBackend) Authenticate(apiKey string) error {
	ret := _m.Called(apiKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(apiKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
