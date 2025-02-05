package main

import (
	"context"
	"dagger/demo/internal/dagger"
)

type Demo struct{}

// Automate a Go programming task with a LLM
// The input is a prompt and an optional starting point.
// The output is the generated / modified source code in a containerized dev environment
func (m *Demo) GoProgrammer(ctx context.Context,
	// A starting point for the coder's workspace.
	// Defaults to an empty directory
	// +optional
	start *dagger.Directory,
	// A description of the Go programming task to perform
	task string,
) *dagger.Container {
	workspace := dag.Workspace(dagger.WorkspaceOpts{
		Start:   start,
		Checker: dag.Go(dag.Directory()).Base().WithDefaultArgs([]string{"go", "build", "./..."}),
	})
	result := dag.
		Llm().
		WithWorkspace(workspace).
		WithPrompt("You are a Go programmer. Use your workspace to accomplish the given task. Check your work with the 'check' tool\n<task>\n" + task + "\n</task>").
		Workspace().
		Dir()
	return dag.Container().
		From("golang").
		WithDirectory(".", result)
}
