package multi_work

import "context"

type Request interface {
	Value() int
}

type Response interface {
	Value() int
}

type WorkFunc func(ctx context.Context, req Request) (Response, error)
