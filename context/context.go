package context

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

type emptyCtx int

func (emptyCtx) Deadline() (deadline time.Time, ok bool) { return }
func (emptyCtx) Done() <-chan struct{}                   { return nil }
func (emptyCtx) Err() error                              { return nil }
func (emptyCtx) Value(key interface{}) interface{}       { return nil }

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func Background() Context { return background }
func TODO() Context       { return todo }

type CancelFunc func()

type cancelCtx struct {
	Context
	done chan struct{}
	err  error
	mu   sync.Mutex
}

var Canceled = errors.New("context cancelled")

func WithCancel(parent Context) (Context, CancelFunc) {
	ctx := &cancelCtx{
		Context: parent,
		done:    make(chan struct{}),
	}

	cancel := func() { ctx.cancel(Canceled) }

	go func() {
		select {
		case <-parent.Done():
			ctx.cancel(parent.Err())
		case <-ctx.done:
		}
	}()

	return ctx, cancel
}

func (ctx *cancelCtx) Done() <-chan struct{} {
	return ctx.done
}

func (ctx *cancelCtx) Err() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.err
}

func (ctx *cancelCtx) cancel(err error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.err != nil {
		return // already canceled
	}

	ctx.err = err
	close(ctx.done)
}

var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string { return "context deadline exceeded" }

func (deadlineExceededError) Timeout() bool { return true }

func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc) {
	ctx, cancel := WithCancel(parent)

	time.AfterFunc(deadline.Sub(time.Now()), func() {
		ctx.(*cancelCtx).cancel(DeadlineExceeded)
	})

	return ctx, cancel
}

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

func WithValue(parent Context, key, val interface{}) Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	return &valueCtx{parent, key, val}
}

type valueCtx struct {
	Context
	key, val interface{}
}

func (ctx *valueCtx) Value(key interface{}) interface{} {
	if ctx.key == key {
		return ctx.val
	}
	return ctx.Context.Value(key)
}
