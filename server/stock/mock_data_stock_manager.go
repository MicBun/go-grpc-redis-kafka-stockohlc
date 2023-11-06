// Code generated by mockery. DO NOT EDIT.

package stock

import (
	context "context"

	pb "github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	mock "github.com/stretchr/testify/mock"
)

// MockDataStockManager is an autogenerated mock type for the DataStockManager type
type MockDataStockManager struct {
	mock.Mock
}

type MockDataStockManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDataStockManager) EXPECT() *MockDataStockManager_Expecter {
	return &MockDataStockManager_Expecter{mock: &_m.Mock}
}

// GetOneSummary provides a mock function with given fields: ctx, in
func (_m *MockDataStockManager) GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	ret := _m.Called(ctx, in)

	var r0 *pb.Stock
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *pb.GetOneSummaryRequest) (*pb.Stock, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *pb.GetOneSummaryRequest) *pb.Stock); ok {
		r0 = rf(ctx, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.Stock)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *pb.GetOneSummaryRequest) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDataStockManager_GetOneSummary_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOneSummary'
type MockDataStockManager_GetOneSummary_Call struct {
	*mock.Call
}

// GetOneSummary is a helper method to define mock.On call
//   - ctx context.Context
//   - in *pb.GetOneSummaryRequest
func (_e *MockDataStockManager_Expecter) GetOneSummary(ctx interface{}, in interface{}) *MockDataStockManager_GetOneSummary_Call {
	return &MockDataStockManager_GetOneSummary_Call{Call: _e.mock.On("GetOneSummary", ctx, in)}
}

func (_c *MockDataStockManager_GetOneSummary_Call) Run(run func(ctx context.Context, in *pb.GetOneSummaryRequest)) *MockDataStockManager_GetOneSummary_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*pb.GetOneSummaryRequest))
	})
	return _c
}

func (_c *MockDataStockManager_GetOneSummary_Call) Return(_a0 *pb.Stock, _a1 error) *MockDataStockManager_GetOneSummary_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDataStockManager_GetOneSummary_Call) RunAndReturn(run func(context.Context, *pb.GetOneSummaryRequest) (*pb.Stock, error)) *MockDataStockManager_GetOneSummary_Call {
	_c.Call.Return(run)
	return _c
}

// LoadInitialData provides a mock function with given fields: ctx
func (_m *MockDataStockManager) LoadInitialData(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataStockManager_LoadInitialData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadInitialData'
type MockDataStockManager_LoadInitialData_Call struct {
	*mock.Call
}

// LoadInitialData is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockDataStockManager_Expecter) LoadInitialData(ctx interface{}) *MockDataStockManager_LoadInitialData_Call {
	return &MockDataStockManager_LoadInitialData_Call{Call: _e.mock.On("LoadInitialData", ctx)}
}

func (_c *MockDataStockManager_LoadInitialData_Call) Run(run func(ctx context.Context)) *MockDataStockManager_LoadInitialData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockDataStockManager_LoadInitialData_Call) Return(_a0 error) *MockDataStockManager_LoadInitialData_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataStockManager_LoadInitialData_Call) RunAndReturn(run func(context.Context) error) *MockDataStockManager_LoadInitialData_Call {
	_c.Call.Return(run)
	return _c
}

// SetMapStockData provides a mock function with given fields: ctx, mapStockData
func (_m *MockDataStockManager) SetMapStockData(ctx context.Context, mapStockData map[string]*pb.Stock) error {
	ret := _m.Called(ctx, mapStockData)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, map[string]*pb.Stock) error); ok {
		r0 = rf(ctx, mapStockData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataStockManager_SetMapStockData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetMapStockData'
type MockDataStockManager_SetMapStockData_Call struct {
	*mock.Call
}

// SetMapStockData is a helper method to define mock.On call
//   - ctx context.Context
//   - mapStockData map[string]*pb.Stock
func (_e *MockDataStockManager_Expecter) SetMapStockData(ctx interface{}, mapStockData interface{}) *MockDataStockManager_SetMapStockData_Call {
	return &MockDataStockManager_SetMapStockData_Call{Call: _e.mock.On("SetMapStockData", ctx, mapStockData)}
}

func (_c *MockDataStockManager_SetMapStockData_Call) Run(run func(ctx context.Context, mapStockData map[string]*pb.Stock)) *MockDataStockManager_SetMapStockData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(map[string]*pb.Stock))
	})
	return _c
}

func (_c *MockDataStockManager_SetMapStockData_Call) Return(_a0 error) *MockDataStockManager_SetMapStockData_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataStockManager_SetMapStockData_Call) RunAndReturn(run func(context.Context, map[string]*pb.Stock) error) *MockDataStockManager_SetMapStockData_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateStockOnFileCreate provides a mock function with given fields: ctx, createdFileName
func (_m *MockDataStockManager) UpdateStockOnFileCreate(ctx context.Context, createdFileName string) error {
	ret := _m.Called(ctx, createdFileName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, createdFileName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataStockManager_UpdateStockOnFileCreate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateStockOnFileCreate'
type MockDataStockManager_UpdateStockOnFileCreate_Call struct {
	*mock.Call
}

// UpdateStockOnFileCreate is a helper method to define mock.On call
//   - ctx context.Context
//   - createdFileName string
func (_e *MockDataStockManager_Expecter) UpdateStockOnFileCreate(ctx interface{}, createdFileName interface{}) *MockDataStockManager_UpdateStockOnFileCreate_Call {
	return &MockDataStockManager_UpdateStockOnFileCreate_Call{Call: _e.mock.On("UpdateStockOnFileCreate", ctx, createdFileName)}
}

func (_c *MockDataStockManager_UpdateStockOnFileCreate_Call) Run(run func(ctx context.Context, createdFileName string)) *MockDataStockManager_UpdateStockOnFileCreate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockDataStockManager_UpdateStockOnFileCreate_Call) Return(_a0 error) *MockDataStockManager_UpdateStockOnFileCreate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataStockManager_UpdateStockOnFileCreate_Call) RunAndReturn(run func(context.Context, string) error) *MockDataStockManager_UpdateStockOnFileCreate_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateStockOnFileUpdate provides a mock function with given fields: ctx, updatedFileName
func (_m *MockDataStockManager) UpdateStockOnFileUpdate(ctx context.Context, updatedFileName string) error {
	ret := _m.Called(ctx, updatedFileName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, updatedFileName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockDataStockManager_UpdateStockOnFileUpdate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateStockOnFileUpdate'
type MockDataStockManager_UpdateStockOnFileUpdate_Call struct {
	*mock.Call
}

// UpdateStockOnFileUpdate is a helper method to define mock.On call
//   - ctx context.Context
//   - updatedFileName string
func (_e *MockDataStockManager_Expecter) UpdateStockOnFileUpdate(ctx interface{}, updatedFileName interface{}) *MockDataStockManager_UpdateStockOnFileUpdate_Call {
	return &MockDataStockManager_UpdateStockOnFileUpdate_Call{Call: _e.mock.On("UpdateStockOnFileUpdate", ctx, updatedFileName)}
}

func (_c *MockDataStockManager_UpdateStockOnFileUpdate_Call) Run(run func(ctx context.Context, updatedFileName string)) *MockDataStockManager_UpdateStockOnFileUpdate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockDataStockManager_UpdateStockOnFileUpdate_Call) Return(_a0 error) *MockDataStockManager_UpdateStockOnFileUpdate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDataStockManager_UpdateStockOnFileUpdate_Call) RunAndReturn(run func(context.Context, string) error) *MockDataStockManager_UpdateStockOnFileUpdate_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDataStockManager creates a new instance of MockDataStockManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDataStockManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDataStockManager {
	mock := &MockDataStockManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
