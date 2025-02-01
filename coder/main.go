package main

import (
	"context"
	"dagger/coder/internal/dagger"
)

func New(
	// A base system for the dev environment
	// +optional
	env *dagger.Container,
	// An initial workspace
	// +optional
	workspace *dagger.Directory,
) Coder {
	if env == nil {
		env = dag.Container()
	}
	if workspace == nil {
		workspace = dag.Directory()
	}
	return Coder{
		Workspace: workspace,
		Env:       env,
	}.Checkpoint("initial checkpoint")
}

type Coder struct {
	// A history of all checkpoints of the workspace so far
	Checkpoints []*Checkpoint
	// The current workspace
	Workspace *dagger.Directory
	// The current dev environment
	Env *dagger.Container
}

type Checkpoint struct {
	// The state of the workspace at the time of checkpoint
	Workspace *dagger.Directory
	// Description of changes since the previous checkpoint
	Message string
}

// Checkpoint the current workspace, with a message explaining changes since the previous checkpoint
func (c Coder) Checkpoint(
	// The message explaining the difference since the previous checkpoint
	message string,
) Coder {
	c.Checkpoints = append(c.Checkpoints, &Checkpoint{
		Workspace: c.Workspace,
		Message:   message,
	})
	return c
}

// Returns the contents of the last checkpoint, without restoring it
func (c Coder) LastCheckpoint() *Checkpoint {
	return c.Checkpoints[len(c.Checkpoints)-1]
}

// Setup the dev environment as specified
func (c Coder) SetupEnv(
	// A human-readable spec describing how to configure the environment
	// <example>A Go dev environment with a glibc-based distro with the latest version of Go, imagemagick installed, libffmpeg, and nodejs</example>
	spec string,
) *dagger.Container {
	return c.Env.
		Agent().
		Please("Configure the container as follows:\n" + spec).
		State()
}

// Return all changes to the workspace since the last checkpoint, in standard diff format
func (c Coder) Diff(ctx context.Context) (string, error) {
	return dag.Container().
		From("alpine").
		WithWorkdir("/workspace").
		WithMountedDirectory("previous", c.LastCheckpoint().Workspace).
		WithMountedDirectory("current", c.Workspace).
		Terminal().
		WithExec([]string{"diff", "-r", "./previous", "./current"}).
		Stdout(ctx)
}

// Write to a file in the workspace
func (c Coder) WriteFile(path string, contents string) Coder {
	c.Workspace = c.Workspace.WithNewFile(path, contents)
	return c
}

// Read the contents of a file in thw workspace
func (c Coder) ReadFile(ctx context.Context, path string) (string, error) {
	// NOTE: just a wrapper on top of dagger API
	// we simulate a simple stateful API on top of stateless primitives
	// we have full control over how constrained, or how open-ended
	// the agent's environemnt is
	return c.Workspace.File(path).Contents(ctx)
}

// Find files in the workspace
// Example patterns:
// <example>**</example>
// <example>*.gif
func (c Coder) FindFiles(
	ctx context.Context,
	// Include filenames matching the patterns. Uses the buildkit pattern format (slightly different from the gitignore format)
	// To match all files: "**"
	pattern string,
) ([]string, error) {
	return c.Workspace.Glob(ctx, pattern)
}
