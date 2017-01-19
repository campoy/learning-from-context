package context

import "time"

var (
	Canceled         error
	DeadlineExceeded error
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

func TODO() Context       { return nil }
func Background() Context { return nil }

type CancelFunc func()

func WithCancel(parent Context) (Context, CancelFunc) {
	return nil, nil
}

func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc) {
	return nil, nil
}

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return nil, nil
}

func WithValue(parent Context, key, val interface{}) Context {
	return nil
}
