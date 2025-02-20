package main

import (
	"context"
	"dagger/go-checker/internal/dagger"
	"errors"
)

type GoChecker struct {
	Dir *dagger.Directory
}

func (checker GoChecker) WithDirectory(dir *dagger.Directory) GoChecker {
	checker.Dir = dir
	return checker
}

func (checker GoChecker) Check(ctx context.Context) error {
	execOpts := dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}
	check := dag.Go(checker.Dir).
		Env().
		WithExec([]string{"sh", "-c", "go mod tidy && go build ./..."}, execOpts)
	code, err := check.ExitCode(ctx)
	if err != nil {
		return err
	}
	if code == 0 {
		return nil
	}
	stderr, err := check.Stderr(ctx)
	if err != nil {
		return err
	}
	return errors.New(stderr)
}
