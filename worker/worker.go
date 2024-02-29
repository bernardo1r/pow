package worker

import (
	"context"

	"github.com/bernardo1r/pow"

	"golang.org/x/sync/errgroup"
)

func work(ctx context.Context, results chan<- *pow.Result, digest []byte, target int) (*pow.Result, error) {
	p, err := pow.New(digest)
	if err != nil {
		return nil, err
	}
	highest := p.Result()
	results <- highest

	if err != nil {
		return nil, err
	}

	var res *pow.Result
	for {
		select {
		case <-ctx.Done():
			return highest, nil

		default:
		}

		res, err = p.Redo(res)
		if err != nil {
			return nil, err
		}
		if res.Zeros <= highest.Zeros {
			continue
		}

		highest = res
		results <- highest
		res = nil

		if highest.Zeros >= target {
			return highest, nil
		}
	}
}

func Run(ctx context.Context, nThreads int, results chan<- *pow.Result, digest []byte, target int) (*pow.Result, error) {
	returnChan := make(chan *pow.Result, nThreads)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	for range nThreads {
		group.Go(func() error {
			res, err := work(ctx, results, digest, target)
			cancel()
			if err != nil {
				return err
			}
			returnChan <- res
			return err
		})
	}
	err := group.Wait()
	if err != nil {
		return nil, err
	}
	close(returnChan)

	p, err := pow.New(digest)
	if err != nil {
		return nil, err
	}
	highest := p.Result()
	for res := range returnChan {
		if res.Zeros > highest.Zeros {

			highest = res
		}
	}
	return highest, nil
}
