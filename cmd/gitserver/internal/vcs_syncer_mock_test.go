// Code generated by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file was generated by running `sg generate` (or `go-mockgen`) at the root of
// this repository. To add additional mocks to this or another package, add a new entry
// to the mockgen.yaml file in the root of this repository.

package internal

import (
	"context"
	"os/exec"
	"sync"

	common "github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	api "github.com/sourcegraph/sourcegraph/internal/api"
	vcs "github.com/sourcegraph/sourcegraph/internal/vcs"
)

// MockVCSSyncer is a mock implementation of the VCSSyncer interface (from
// the package github.com/sourcegraph/sourcegraph/cmd/gitserver/internal)
// used for unit testing.
type MockVCSSyncer struct {
	// CloneCommandFunc is an instance of a mock function object controlling
	// the behavior of the method CloneCommand.
	CloneCommandFunc *VCSSyncerCloneCommandFunc
	// FetchFunc is an instance of a mock function object controlling the
	// behavior of the method Fetch.
	FetchFunc *VCSSyncerFetchFunc
	// IsCloneableFunc is an instance of a mock function object controlling
	// the behavior of the method IsCloneable.
	IsCloneableFunc *VCSSyncerIsCloneableFunc
	// RemoteShowCommandFunc is an instance of a mock function object
	// controlling the behavior of the method RemoteShowCommand.
	RemoteShowCommandFunc *VCSSyncerRemoteShowCommandFunc
	// TypeFunc is an instance of a mock function object controlling the
	// behavior of the method Type.
	TypeFunc *VCSSyncerTypeFunc
}

// NewMockVCSSyncer creates a new mock of the VCSSyncer interface. All
// methods return zero values for all results, unless overwritten.
func NewMockVCSSyncer() *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommandFunc: &VCSSyncerCloneCommandFunc{
			defaultHook: func(context.Context, *vcs.URL, string) (r0 *exec.Cmd, r1 error) {
				return
			},
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defaultHook: func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) (r0 []byte, r1 error) {
				return
			},
		},
		IsCloneableFunc: &VCSSyncerIsCloneableFunc{
			defaultHook: func(context.Context, api.RepoName, *vcs.URL) (r0 error) {
				return
			},
		},
		RemoteShowCommandFunc: &VCSSyncerRemoteShowCommandFunc{
			defaultHook: func(context.Context, *vcs.URL) (r0 *exec.Cmd, r1 error) {
				return
			},
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defaultHook: func() (r0 string) {
				return
			},
		},
	}
}

// NewStrictMockVCSSyncer creates a new mock of the VCSSyncer interface. All
// methods panic on invocation, unless overwritten.
func NewStrictMockVCSSyncer() *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommandFunc: &VCSSyncerCloneCommandFunc{
			defaultHook: func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
				panic("unexpected invocation of MockVCSSyncer.CloneCommand")
			},
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defaultHook: func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error) {
				panic("unexpected invocation of MockVCSSyncer.Fetch")
			},
		},
		IsCloneableFunc: &VCSSyncerIsCloneableFunc{
			defaultHook: func(context.Context, api.RepoName, *vcs.URL) error {
				panic("unexpected invocation of MockVCSSyncer.IsCloneable")
			},
		},
		RemoteShowCommandFunc: &VCSSyncerRemoteShowCommandFunc{
			defaultHook: func(context.Context, *vcs.URL) (*exec.Cmd, error) {
				panic("unexpected invocation of MockVCSSyncer.RemoteShowCommand")
			},
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defaultHook: func() string {
				panic("unexpected invocation of MockVCSSyncer.Type")
			},
		},
	}
}

// NewMockVCSSyncerFrom creates a new mock of the MockVCSSyncer interface.
// All methods delegate to the given implementation, unless overwritten.
func NewMockVCSSyncerFrom(i VCSSyncer) *MockVCSSyncer {
	return &MockVCSSyncer{
		CloneCommandFunc: &VCSSyncerCloneCommandFunc{
			defaultHook: i.CloneCommand,
		},
		FetchFunc: &VCSSyncerFetchFunc{
			defaultHook: i.Fetch,
		},
		IsCloneableFunc: &VCSSyncerIsCloneableFunc{
			defaultHook: i.IsCloneable,
		},
		RemoteShowCommandFunc: &VCSSyncerRemoteShowCommandFunc{
			defaultHook: i.RemoteShowCommand,
		},
		TypeFunc: &VCSSyncerTypeFunc{
			defaultHook: i.Type,
		},
	}
}

// VCSSyncerCloneCommandFunc describes the behavior when the CloneCommand
// method of the parent MockVCSSyncer instance is invoked.
type VCSSyncerCloneCommandFunc struct {
	defaultHook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)
	hooks       []func(context.Context, *vcs.URL, string) (*exec.Cmd, error)
	history     []VCSSyncerCloneCommandFuncCall
	mutex       sync.Mutex
}

// CloneCommand delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockVCSSyncer) CloneCommand(v0 context.Context, v1 *vcs.URL, v2 string) (*exec.Cmd, error) {
	r0, r1 := m.CloneCommandFunc.nextHook()(v0, v1, v2)
	m.CloneCommandFunc.appendCall(VCSSyncerCloneCommandFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the CloneCommand method
// of the parent MockVCSSyncer instance is invoked and the hook queue is
// empty.
func (f *VCSSyncerCloneCommandFunc) SetDefaultHook(hook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CloneCommand method of the parent MockVCSSyncer instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *VCSSyncerCloneCommandFunc) PushHook(hook func(context.Context, *vcs.URL, string) (*exec.Cmd, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *VCSSyncerCloneCommandFunc) SetDefaultReturn(r0 *exec.Cmd, r1 error) {
	f.SetDefaultHook(func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *VCSSyncerCloneCommandFunc) PushReturn(r0 *exec.Cmd, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
		return r0, r1
	})
}

func (f *VCSSyncerCloneCommandFunc) nextHook() func(context.Context, *vcs.URL, string) (*exec.Cmd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerCloneCommandFunc) appendCall(r0 VCSSyncerCloneCommandFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of VCSSyncerCloneCommandFuncCall objects
// describing the invocations of this function.
func (f *VCSSyncerCloneCommandFunc) History() []VCSSyncerCloneCommandFuncCall {
	f.mutex.Lock()
	history := make([]VCSSyncerCloneCommandFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerCloneCommandFuncCall is an object that describes an invocation
// of method CloneCommand on an instance of MockVCSSyncer.
type VCSSyncerCloneCommandFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *vcs.URL
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *exec.Cmd
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c VCSSyncerCloneCommandFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c VCSSyncerCloneCommandFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// VCSSyncerFetchFunc describes the behavior when the Fetch method of the
// parent MockVCSSyncer instance is invoked.
type VCSSyncerFetchFunc struct {
	defaultHook func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error)
	hooks       []func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error)
	history     []VCSSyncerFetchFuncCall
	mutex       sync.Mutex
}

// Fetch delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockVCSSyncer) Fetch(v0 context.Context, v1 *vcs.URL, v2 api.RepoName, v3 common.GitDir, v4 string) ([]byte, error) {
	r0, r1 := m.FetchFunc.nextHook()(v0, v1, v2, v3, v4)
	m.FetchFunc.appendCall(VCSSyncerFetchFuncCall{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Fetch method of the
// parent MockVCSSyncer instance is invoked and the hook queue is empty.
func (f *VCSSyncerFetchFunc) SetDefaultHook(hook func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Fetch method of the parent MockVCSSyncer instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *VCSSyncerFetchFunc) PushHook(hook func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *VCSSyncerFetchFunc) SetDefaultReturn(r0 []byte, r1 error) {
	f.SetDefaultHook(func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *VCSSyncerFetchFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error) {
		return r0, r1
	})
}

func (f *VCSSyncerFetchFunc) nextHook() func(context.Context, *vcs.URL, api.RepoName, common.GitDir, string) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerFetchFunc) appendCall(r0 VCSSyncerFetchFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of VCSSyncerFetchFuncCall objects describing
// the invocations of this function.
func (f *VCSSyncerFetchFunc) History() []VCSSyncerFetchFuncCall {
	f.mutex.Lock()
	history := make([]VCSSyncerFetchFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerFetchFuncCall is an object that describes an invocation of
// method Fetch on an instance of MockVCSSyncer.
type VCSSyncerFetchFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *vcs.URL
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 api.RepoName
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 common.GitDir
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []byte
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c VCSSyncerFetchFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c VCSSyncerFetchFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// VCSSyncerIsCloneableFunc describes the behavior when the IsCloneable
// method of the parent MockVCSSyncer instance is invoked.
type VCSSyncerIsCloneableFunc struct {
	defaultHook func(context.Context, api.RepoName, *vcs.URL) error
	hooks       []func(context.Context, api.RepoName, *vcs.URL) error
	history     []VCSSyncerIsCloneableFuncCall
	mutex       sync.Mutex
}

// IsCloneable delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockVCSSyncer) IsCloneable(v0 context.Context, v1 api.RepoName, v2 *vcs.URL) error {
	r0 := m.IsCloneableFunc.nextHook()(v0, v1, v2)
	m.IsCloneableFunc.appendCall(VCSSyncerIsCloneableFuncCall{v0, v1, v2, r0})
	return r0
}

// SetDefaultHook sets function that is called when the IsCloneable method
// of the parent MockVCSSyncer instance is invoked and the hook queue is
// empty.
func (f *VCSSyncerIsCloneableFunc) SetDefaultHook(hook func(context.Context, api.RepoName, *vcs.URL) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// IsCloneable method of the parent MockVCSSyncer instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *VCSSyncerIsCloneableFunc) PushHook(hook func(context.Context, api.RepoName, *vcs.URL) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *VCSSyncerIsCloneableFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, api.RepoName, *vcs.URL) error {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *VCSSyncerIsCloneableFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, api.RepoName, *vcs.URL) error {
		return r0
	})
}

func (f *VCSSyncerIsCloneableFunc) nextHook() func(context.Context, api.RepoName, *vcs.URL) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerIsCloneableFunc) appendCall(r0 VCSSyncerIsCloneableFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of VCSSyncerIsCloneableFuncCall objects
// describing the invocations of this function.
func (f *VCSSyncerIsCloneableFunc) History() []VCSSyncerIsCloneableFuncCall {
	f.mutex.Lock()
	history := make([]VCSSyncerIsCloneableFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerIsCloneableFuncCall is an object that describes an invocation of
// method IsCloneable on an instance of MockVCSSyncer.
type VCSSyncerIsCloneableFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 api.RepoName
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 *vcs.URL
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c VCSSyncerIsCloneableFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c VCSSyncerIsCloneableFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// VCSSyncerRemoteShowCommandFunc describes the behavior when the
// RemoteShowCommand method of the parent MockVCSSyncer instance is invoked.
type VCSSyncerRemoteShowCommandFunc struct {
	defaultHook func(context.Context, *vcs.URL) (*exec.Cmd, error)
	hooks       []func(context.Context, *vcs.URL) (*exec.Cmd, error)
	history     []VCSSyncerRemoteShowCommandFuncCall
	mutex       sync.Mutex
}

// RemoteShowCommand delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockVCSSyncer) RemoteShowCommand(v0 context.Context, v1 *vcs.URL) (*exec.Cmd, error) {
	r0, r1 := m.RemoteShowCommandFunc.nextHook()(v0, v1)
	m.RemoteShowCommandFunc.appendCall(VCSSyncerRemoteShowCommandFuncCall{v0, v1, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the RemoteShowCommand
// method of the parent MockVCSSyncer instance is invoked and the hook queue
// is empty.
func (f *VCSSyncerRemoteShowCommandFunc) SetDefaultHook(hook func(context.Context, *vcs.URL) (*exec.Cmd, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// RemoteShowCommand method of the parent MockVCSSyncer instance invokes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *VCSSyncerRemoteShowCommandFunc) PushHook(hook func(context.Context, *vcs.URL) (*exec.Cmd, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *VCSSyncerRemoteShowCommandFunc) SetDefaultReturn(r0 *exec.Cmd, r1 error) {
	f.SetDefaultHook(func(context.Context, *vcs.URL) (*exec.Cmd, error) {
		return r0, r1
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *VCSSyncerRemoteShowCommandFunc) PushReturn(r0 *exec.Cmd, r1 error) {
	f.PushHook(func(context.Context, *vcs.URL) (*exec.Cmd, error) {
		return r0, r1
	})
}

func (f *VCSSyncerRemoteShowCommandFunc) nextHook() func(context.Context, *vcs.URL) (*exec.Cmd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerRemoteShowCommandFunc) appendCall(r0 VCSSyncerRemoteShowCommandFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of VCSSyncerRemoteShowCommandFuncCall objects
// describing the invocations of this function.
func (f *VCSSyncerRemoteShowCommandFunc) History() []VCSSyncerRemoteShowCommandFuncCall {
	f.mutex.Lock()
	history := make([]VCSSyncerRemoteShowCommandFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerRemoteShowCommandFuncCall is an object that describes an
// invocation of method RemoteShowCommand on an instance of MockVCSSyncer.
type VCSSyncerRemoteShowCommandFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 *vcs.URL
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 *exec.Cmd
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c VCSSyncerRemoteShowCommandFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c VCSSyncerRemoteShowCommandFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// VCSSyncerTypeFunc describes the behavior when the Type method of the
// parent MockVCSSyncer instance is invoked.
type VCSSyncerTypeFunc struct {
	defaultHook func() string
	hooks       []func() string
	history     []VCSSyncerTypeFuncCall
	mutex       sync.Mutex
}

// Type delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockVCSSyncer) Type() string {
	r0 := m.TypeFunc.nextHook()()
	m.TypeFunc.appendCall(VCSSyncerTypeFuncCall{r0})
	return r0
}

// SetDefaultHook sets function that is called when the Type method of the
// parent MockVCSSyncer instance is invoked and the hook queue is empty.
func (f *VCSSyncerTypeFunc) SetDefaultHook(hook func() string) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Type method of the parent MockVCSSyncer instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *VCSSyncerTypeFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultHook with a function that returns the
// given values.
func (f *VCSSyncerTypeFunc) SetDefaultReturn(r0 string) {
	f.SetDefaultHook(func() string {
		return r0
	})
}

// PushReturn calls PushHook with a function that returns the given values.
func (f *VCSSyncerTypeFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *VCSSyncerTypeFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *VCSSyncerTypeFunc) appendCall(r0 VCSSyncerTypeFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of VCSSyncerTypeFuncCall objects describing
// the invocations of this function.
func (f *VCSSyncerTypeFunc) History() []VCSSyncerTypeFuncCall {
	f.mutex.Lock()
	history := make([]VCSSyncerTypeFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// VCSSyncerTypeFuncCall is an object that describes an invocation of method
// Type on an instance of MockVCSSyncer.
type VCSSyncerTypeFuncCall struct {
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c VCSSyncerTypeFuncCall) Args() []interface{} {
	return []interface{}{}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c VCSSyncerTypeFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
