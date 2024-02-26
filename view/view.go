package view

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/bernardo1r/pow"
)

func display(result *pow.Result) {
	fmt.Println("-----------------------------")
	fmt.Printf("Number of zeroes: %v\n", result.Zeros)
	fmt.Printf("Challenge:       %v\n", hex.EncodeToString(result.Challenge))
	fmt.Printf("Final digest:    %v\n", hex.EncodeToString(result.Digest))
}

func print(ctx context.Context, results <-chan *pow.Result) {
	highest := new(pow.Result)
	for {
		var res *pow.Result
		select {
		case <-ctx.Done():
			return

		case res = <-results:
		}

		if res.Zeros <= highest.Zeros {
			continue
		}

		highest = res
		display(highest)
	}
}

func Run(ctx context.Context, results <-chan *pow.Result) {
	print(ctx, results)
}
