// A toy workspace for editing and building go programs
package main

import (
	"context"
	"dagger/toy-workspace/internal/dagger"
)

func New() ToyWorkspace {
	return ToyWorkspace{
		Container: dag.Container().
			From("golang").
			WithDefaultTerminalCmd([]string{"/bin/bash"}).
			WithMountedCache("/go/pkg/mod", dag.CacheVolume("go_mod_cache")).
			WithWorkdir("/app"),
	}
}

type ToyWorkspace struct {
	// The workspace container.
	// +internal-use-only
	Container *dagger.Container
}

// Read a file
func (w *ToyWorkspace) Read(
	ctx context.Context,
	// The path of the file
	path string) (string, error) {
	return w.Container.File(path).Contents(ctx)
}

// Write a file
func (w *ToyWorkspace) Write(
	// The path of the file
	path string,
	// The content to write
	content string,
) *ToyWorkspace {
	w.Container = w.Container.WithNewFile(path, content)
	return w
}

// Build the code at the current directory in the workspace
func (w *ToyWorkspace) Build(ctx context.Context) (string, error) {
	// We just execute "go build" in the container,
	buildCommand := []string{"go", "build", "./..."}
	return w.Container.
		WithExec(buildCommand, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		Stderr(ctx)
}
