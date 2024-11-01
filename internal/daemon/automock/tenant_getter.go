// Code generated by mockery v2.10.0. DO NOT EDIT.

package automock

import (
	repo "github.com/mszostok/job-runner/pkg/job/repo"
	mock "github.com/stretchr/testify/mock"
)

// TenantGetter is an autogenerated mock type for the TenantGetter type
type TenantGetter struct {
	mock.Mock
}

type TenantGetter_Expecter struct {
	mock *mock.Mock
}

func (_m *TenantGetter) EXPECT() *TenantGetter_Expecter {
	return &TenantGetter_Expecter{mock: &_m.Mock}
}

// GetJobTenant provides a mock function with given fields: in
func (_m *TenantGetter) GetJobTenant(in repo.GetJobTenantInput) (repo.GetJobTenantOutput, error) {
	ret := _m.Called(in)

	var r0 repo.GetJobTenantOutput
	if rf, ok := ret.Get(0).(func(repo.GetJobTenantInput) repo.GetJobTenantOutput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(repo.GetJobTenantOutput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(repo.GetJobTenantInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TenantGetter_GetJobTenant_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJobTenant'
type TenantGetter_GetJobTenant_Call struct {
	*mock.Call
}

// GetJobTenant is a helper method to define mock.On call
//  - in repo.GetJobTenantInput
func (_e *TenantGetter_Expecter) GetJobTenant(in interface{}) *TenantGetter_GetJobTenant_Call {
	return &TenantGetter_GetJobTenant_Call{Call: _e.mock.On("GetJobTenant", in)}
}

func (_c *TenantGetter_GetJobTenant_Call) Run(run func(in repo.GetJobTenantInput)) *TenantGetter_GetJobTenant_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(repo.GetJobTenantInput))
	})
	return _c
}

func (_c *TenantGetter_GetJobTenant_Call) Return(_a0 repo.GetJobTenantOutput, _a1 error) *TenantGetter_GetJobTenant_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
