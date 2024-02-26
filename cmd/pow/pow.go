package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"

	"github.com/bernardo1r/pow"
	"github.com/bernardo1r/pow/internal/save"
	"github.com/bernardo1r/pow/view"
	"github.com/bernardo1r/pow/worker"

	"golang.org/x/sync/errgroup"
)

const ouputFile = "result.json"

const usage = `Usage: pow [OPTION] [INPUTFILE]
Proof of Work toy program that performs PoW on
a file using SHA3-512.

Default -z option is to find a digest made only
of zeros. Default -t option is 1.

Options:

    -z  target number of zeros in the beginning 
        of digest
    -t  number of threads`

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context, digest []byte, target int, nThreads int) (*pow.Result, error) {
	results := make(chan *pow.Result, nThreads)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	var res *pow.Result
	group.Go(func() error {
		defer cancel()
		var err error
		res, err = worker.Run(ctx, nThreads, results, digest, target)
		return err

	})
	group.Go(func() error {
		view.Run(ctx, results)
		return nil
	})

	err := group.Wait()
	return res, err
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Profile-guided optimizations
	pfile, err := os.Create("default.pgo")
	checkError(err)
	defer pfile.Close()
	err = pprof.StartCPUProfile(pfile)
	checkError(err)
	defer pprof.StopCPUProfile()

	var target, nthreads int
	flag.IntVar(&target, "z", 512, "target number of zeroes")
	flag.IntVar(&nthreads, "t", 1, "number of threads")
	flag.Parse()

	if target < 1 || target > 512 {
		log.Fatalln("pow: invalid target number, must be 1 <= target <= 512")
	}
	if nthreads < 1 {
		log.Fatalln("pow: invalid number of threads, must be 1 <= threads")
	}

	if n := flag.NArg(); n == 0 {
		log.Fatalln("pow: no input file specified")
	} else if n > 1 {
		log.Fatalln("pow: only one file can be provided at a time")
	}

	inputfile := filepath.Clean(flag.Arg(0))
	digest, err := pow.DigestFile(inputfile)
	checkError(err)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	res, err := run(ctx, digest, target, nthreads)
	stop()
	checkError(err)

	entry := save.Entry{
		FileDigest: digest,
		Challenge:  res.Challenge,
		PowDigest:  res.Digest,
		Zeros:      res.Zeros,
	}
	err = save.Save(ouputFile, inputfile, &entry)
	checkError(err)
}
