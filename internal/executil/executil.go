package executil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

var ErrTimeout = errors.New("command timed out")

type RunOptions struct {
	Timeout      time.Duration
	Retries      int
	RetryDelay   time.Duration
	FilterStdout func(string) string
	FilterStderr func(string) string
	Verbose      bool
}

func RunWithCapture(ctx context.Context, cmd *exec.Cmd, opts RunOptions) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= opts.Retries; attempt++ {
		runCtx := ctx
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			runCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}

		cmdCtx := exec.CommandContext(runCtx, cmd.Path, cmd.Args[1:]...)
		cmdCtx.Env = cmd.Env

		stdoutBuf := &bytes.Buffer{}
		stderrBuf := &bytes.Buffer{}
		cmdCtx.Stdout = io.MultiWriter(stdoutBuf)
		cmdCtx.Stderr = io.MultiWriter(stderrBuf)

		err := cmdCtx.Run()
		outStr := stdoutBuf.String()
		errStr := stderrBuf.String()

		if opts.FilterStdout != nil {
			outStr = opts.FilterStdout(outStr)
		}
		if opts.FilterStderr != nil {
			errStr = opts.FilterStderr(errStr)
		}

		if opts.Verbose {
			if outStr != "" {
				fmt.Fprint(os.Stdout, outStr)
			}
			if errStr != "" {
				fmt.Fprint(os.Stderr, errStr)
			}
		}

		if runCtx.Err() == context.DeadlineExceeded {
			lastErr = fmt.Errorf("%w after %v", ErrTimeout, opts.Timeout)
		} else if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w; stderr=%q; stdout=%q", attempt+1, err, errStr, outStr)
		} else {
			return outStr + errStr, nil
		}

		if attempt < opts.Retries {
			time.Sleep(opts.RetryDelay)
		}
	}
	return "", lastErr
}
