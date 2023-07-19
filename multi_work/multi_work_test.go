package multi_work

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

func TestMultiWork(t *testing.T) {
	const (
		instantly = time.Duration(0)
		oneTime   = time.Millisecond // for clarity, can be changed to 1 second
		twoTime   = oneTime * 2
		threeTime = oneTime * 3
		fourTime  = oneTime * 4
		fiveTime  = oneTime * 5

		// the time interval for which dependent (internal) goroutines are killed
		checkGoroutines = time.Millisecond * 10
	)

	testCases := []struct {
		name     string
		timeout  time.Duration
		fns      []WorkFunc
		wantResp responseChecker
		wantErr  string
	}{
		{
			name: "one worker",
			fns: []WorkFunc{
				workOK(&response{value: 1}, instantly),
			},
			wantResp: &response{value: 1},
		},
		{
			name: "many workers instantly",
			fns: []WorkFunc{
				workOK(&response{value: 1}, instantly),
				workOK(&response{value: 2}, instantly),
				workOK(&response{value: 3}, instantly),
				workOK(&response{value: 4}, instantly),
				workOK(&response{value: 5}, instantly),
			},
			wantResp: &responseNotZero{},
		},
		{
			name: "many workers",
			fns: []WorkFunc{
				workOK(&response{value: 1}, fourTime),
				workOK(&response{value: 2}, fiveTime),
				workOK(&response{value: 3}, threeTime),
				workOK(&response{value: 4}, oneTime),
				workOK(&response{value: 5}, twoTime),
			},
			wantResp: &response{value: 4},
		},
		{
			name: "all erroneous workers",
			fns: []WorkFunc{
				workError(fmt.Errorf("err1"), oneTime),
				workError(fmt.Errorf("err2"), threeTime),
				workError(fmt.Errorf("err3"), twoTime),
			},
			wantErr: "err2",
		},
		{
			name: "at the last moment",
			fns: []WorkFunc{
				workError(fmt.Errorf("err1"), oneTime),
				workOK(&response{value: 1}, fourTime),
				workError(fmt.Errorf("err2"), threeTime),
				workError(fmt.Errorf("err3"), twoTime),
			},
			wantResp: &response{value: 1},
		},
		{
			name: "looong work",
			fns: []WorkFunc{
				workError(fmt.Errorf("err1"), fourTime*10),
				workOK(&response{value: 1}, fourTime),
			},
			wantResp: &response{value: 1},
		},
		{
			name:    "stop long work",
			timeout: twoTime,
			fns: []WorkFunc{
				workError(fmt.Errorf("err1"), fourTime),
				workError(fmt.Errorf("err2"), fourTime),
				workOK(&response{value: 3}, threeTime),
				workError(fmt.Errorf("err4"), fiveTime),
			},
			wantErr: context.DeadlineExceeded.Error(),
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, test.timeout)
				t.Cleanup(func() {
					cancel()
				})
			}

			req := &request{value: 1}

			numGoroutines := runtime.NumGoroutine()
			got, err := MultiWork(ctx, req, test.fns)
			if test.wantErr == "" {
				if !assert.NoError(t, err) {
					return
				}
			} else {
				if !assert.EqualError(t, err, test.wantErr) {
					return
				}
				return
			}
			assert.True(t, test.wantResp.Check(got.Value()))

			time.Sleep(checkGoroutines)
			assert.Equal(t, numGoroutines, runtime.NumGoroutine(), "number of goroutines")
		})
	}
}

type request struct {
	value int
}

func (r *request) Value() int { return r.value }

type responseChecker interface {
	Response
	Check(v int) bool
}

type response struct {
	value int
}

func (r *response) Value() int       { return r.value }
func (r *response) Check(v int) bool { return r.value == v }

type responseNotZero struct{}

func (r *responseNotZero) Value() int       { return 0 }
func (r *responseNotZero) Check(v int) bool { return v != 0 }

func workOK(resp Response, delay time.Duration) func(ctx context.Context, req Request) (Response, error) {
	return func(ctx context.Context, req Request) (Response, error) {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return resp, nil
		}
	}
}

func workError(err error, delay time.Duration) func(ctx context.Context, req Request) (Response, error) {
	return func(ctx context.Context, req Request) (Response, error) {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, err
		}
	}
}
