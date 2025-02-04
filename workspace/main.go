package main

import (
	"context"
	"dagger/workspace/internal/dagger"
)

func New(
	// A builder container, for verifying that the code builds, tests run etc.
	// Container spec:
	// - Workspace will be mounted to container workdir
	// - Container default args will be executed.
	// - Exit code 0 is considered a successful check. Otherwise a failure.
	// +optional
	checker *dagger.Container,
	// Initial state to start the workspace from
	// By default the workspace starts empty
	// +optional
	start *dagger.Directory,
) Workspace {
	if start == nil {
		start = dag.Directory()
	}
	return Workspace{
		Start:   start,
		Dir:     start,
		Checker: checker,
	}
}

// A workspace for editing files and checking the result
type Workspace struct {
	Start   *dagger.Directory // +private
	Dir     *dagger.Directory // +private
	Checker *dagger.Container // +private
}

// Check that the current contents is valid
// This is done by executed an externally-provided checker container with the workspace mounted.
// If there is no checker, the check will always pass
func (s Workspace) Check(ctx context.Context) error {
	if s.Checker == nil {
		return nil
	}
	_, err := s.Checker.
		WithMountedDirectory(".", s.Dir).
		WithExec(nil, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny,
		}).
		Sync(ctx)
	return err
}

// Return all changes to the workspace since the start of the session
func (ws Workspace) Diff(ctx context.Context) (string, error) {
	return base().
		WithWorkdir("/workspace").
		WithMountedDirectory("start", ws.Start).
		WithMountedDirectory("current", ws.Dir).
		WithExec([]string{"diff", "-r", "./start", "./current"}).
		Stdout(ctx)
}

// Reset the workspace to its starting state.
// Warning: this will wipe all changes made during the current session
func (ws Workspace) Reset() Workspace {
	ws.Dir = ws.Start
	return ws
}

// Write to a file in the workspace
func (ws Workspace) Write(
	// The path of the file to write
	path string,
	// The contents to write
	contents string,
) Workspace {
	ws.Dir = ws.Dir.WithNewFile(path, contents)
	return ws
}

// Read the contents of a file in thw workspace
func (ws Workspace) Read(ctx context.Context, path string) (string, error) {
	return ws.Dir.File(path).Contents(ctx)
}

// Remove a file from the workspace
func (ws Workspace) Rm(path string) Workspace {
	ws.Dir = ws.Dir.WithoutFile(path)
	return ws
}

// Remove a directory from the workspace
func (ws Workspace) RmDir(path string) Workspace {
	ws.Dir = ws.Dir.WithoutDirectory(path)
	return ws
}

// List the contents of a directory in the workspace
func (ws Workspace) ListDir(
	ctx context.Context,
	// Path of the target directory
	// +optional
	// +default="/"
	path string,
) ([]string, error) {
	return ws.Dir.Directory(path).Entries(ctx)
}

// Walk all files in the workspace (optionally filtered by a glob pattern), and return their path.
func (ws Workspace) Walk(
	ctx context.Context,
	// A glob pattern to filter files. Only matching files will be included.
	// The glob format is the same as Dockerfile/buildkit
	// +optional
	// +default="**"
	pattern string,
) ([]string, error) {
	return ws.Dir.Glob(ctx, pattern)
}

// A base container for running basic unix utilities with minimal overhead
func base() *dagger.Container {
	digest := "sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099"
	return dag.
		Container().
		From("docker.io/library/alpine:latest@" + digest)
}
