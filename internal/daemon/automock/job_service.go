// Code generated by mockery v2.10.0. DO NOT EDIT.

package automock

import (
	context "context"

	job "github.com/mszostok/job-runner/pkg/job"

	mock "github.com/stretchr/testify/mock"
)

// JobService is an autogenerated mock type for the JobService type
type JobService struct {
	mock.Mock
}

type JobService_Expecter struct {
	mock *mock.Mock
}

func (_m *JobService) EXPECT() *JobService_Expecter {
	return &JobService_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *JobService) Get(_a0 context.Context, _a1 job.GetInput) (*job.GetOutput, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *job.GetOutput
	if rf, ok := ret.Get(0).(func(context.Context, job.GetInput) *job.GetOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.GetOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, job.GetInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// JobService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type JobService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 job.GetInput
func (_e *JobService_Expecter) Get(_a0 interface{}, _a1 interface{}) *JobService_Get_Call {
	return &JobService_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *JobService_Get_Call) Run(run func(_a0 context.Context, _a1 job.GetInput)) *JobService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(job.GetInput))
	})
	return _c
}

func (_c *JobService_Get_Call) Return(_a0 *job.GetOutput, _a1 error) *JobService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Run provides a mock function with given fields: _a0, _a1
func (_m *JobService) Run(_a0 context.Context, _a1 job.RunInput) (*job.RunOutput, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *job.RunOutput
	if rf, ok := ret.Get(0).(func(context.Context, job.RunInput) *job.RunOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.RunOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, job.RunInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// JobService_Run_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Run'
type JobService_Run_Call struct {
	*mock.Call
}

// Run is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 job.RunInput
func (_e *JobService_Expecter) Run(_a0 interface{}, _a1 interface{}) *JobService_Run_Call {
	return &JobService_Run_Call{Call: _e.mock.On("Run", _a0, _a1)}
}

func (_c *JobService_Run_Call) Run(run func(_a0 context.Context, _a1 job.RunInput)) *JobService_Run_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(job.RunInput))
	})
	return _c
}

func (_c *JobService_Run_Call) Return(_a0 *job.RunOutput, _a1 error) *JobService_Run_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Stop provides a mock function with given fields: _a0, _a1
func (_m *JobService) Stop(_a0 context.Context, _a1 job.StopInput) (*job.StopOutput, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *job.StopOutput
	if rf, ok := ret.Get(0).(func(context.Context, job.StopInput) *job.StopOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.StopOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, job.StopInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// JobService_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type JobService_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 job.StopInput
func (_e *JobService_Expecter) Stop(_a0 interface{}, _a1 interface{}) *JobService_Stop_Call {
	return &JobService_Stop_Call{Call: _e.mock.On("Stop", _a0, _a1)}
}

func (_c *JobService_Stop_Call) Run(run func(_a0 context.Context, _a1 job.StopInput)) *JobService_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(job.StopInput))
	})
	return _c
}

func (_c *JobService_Stop_Call) Return(_a0 *job.StopOutput, _a1 error) *JobService_Stop_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// StreamLogs provides a mock function with given fields: _a0, _a1
func (_m *JobService) StreamLogs(_a0 context.Context, _a1 job.StreamLogsInput) (*job.StreamLogsOutput, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *job.StreamLogsOutput
	if rf, ok := ret.Get(0).(func(context.Context, job.StreamLogsInput) *job.StreamLogsOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.StreamLogsOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, job.StreamLogsInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// JobService_StreamLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StreamLogs'
type JobService_StreamLogs_Call struct {
	*mock.Call
}

// StreamLogs is a helper method to define mock.On call
//  - _a0 context.Context
//  - _a1 job.StreamLogsInput
func (_e *JobService_Expecter) StreamLogs(_a0 interface{}, _a1 interface{}) *JobService_StreamLogs_Call {
	return &JobService_StreamLogs_Call{Call: _e.mock.On("StreamLogs", _a0, _a1)}
}

func (_c *JobService_StreamLogs_Call) Run(run func(_a0 context.Context, _a1 job.StreamLogsInput)) *JobService_StreamLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(job.StreamLogsInput))
	})
	return _c
}

func (_c *JobService_StreamLogs_Call) Return(_a0 *job.StreamLogsOutput, _a1 error) *JobService_StreamLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
