package main

import (
	"context"
	"dagger/reviewer/internal/dagger"
	"fmt"
)

type Reviewer struct {
}

func (reviewer *Reviewer) AddReview(
	// The original assignment
	assignment string,
	// The result of the task to be reviewed
	dir *dagger.Directory,
	// The changes that led to the result, in standard diff format
	diff string,
) *dagger.Directory {
	ws := dag.Workspace(dagger.WorkspaceOpts{Start: dir}).
		Write(".review/assignment", assignment).
		Write(".review/diff", diff)
	return dag.Llm().
		WithWorkspace(ws).
		WithPromptFile(dag.CurrentModule().Source().File("prompt.txt")).
		Workspace().
		Dir()
}

// Extract the score from a directory with a review added
func (reviewer *Reviewer) Score(ctx context.Context, dir *dagger.Directory) (int, error) {
	s, err := dir.File(".review/score").Contents(ctx)
	if err != nil {
		return 0, err
	}
	score := 0
	_, err = fmt.Sscanf(s, "%d", &score)
	return score, err
}
