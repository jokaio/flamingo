// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"

import mock "github.com/stretchr/testify/mock"
import url "net/url"
import web "flamingo.me/flamingo/framework/web"

// FormDataDecoder is an autogenerated mock type for the FormDataDecoder type
type FormDataDecoder struct {
	mock.Mock
}

// Decode provides a mock function with given fields: ctx, req, values, formData
func (_m *FormDataDecoder) Decode(ctx context.Context, req *web.Request, values url.Values, formData interface{}) (interface{}, error) {
	ret := _m.Called(ctx, req, values, formData)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(context.Context, *web.Request, url.Values, interface{}) interface{}); ok {
		r0 = rf(ctx, req, values, formData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *web.Request, url.Values, interface{}) error); ok {
		r1 = rf(ctx, req, values, formData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}