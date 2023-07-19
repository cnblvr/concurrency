package multi_work

import (
	"context"
)

func MultiWork(ctx context.Context, req Request, fns []WorkFunc) (Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	respCh, errCh := make(chan Response, 1), make(chan error, 1)
	defer close(respCh)
	defer close(errCh)
	for _, fn := range fns {
		work := fn
		go func() {
			resp, err := work(ctx, req)
			if err != nil {
				errCh <- err
				return
			}
			respCh <- resp
		}()
	}

	var respOut Response
	var errOut error
	// exits from all goroutines are expected
	for i := 0; i < len(fns); i++ {
		select {
		case resp := <-respCh:
			if respOut == nil {
				respOut = resp
				cancel()
			}
		case errOut = <-errCh:
		}
	}
	if respOut != nil {
		return respOut, nil
	}
	return nil, errOut
}
