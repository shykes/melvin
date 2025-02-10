package main

import (
	"context"
	"dagger/demo/internal/dagger"
	"fmt"
	"strings"
)

func New(
	token *dagger.Secret,
	repo string,
	issue int,
) Demo {
	return Demo{
		Token: token,
		Repo:  repo,
		Issue: issue,
	}
}

type Demo struct {
	Token *dagger.Secret
	Repo  string
	Issue int
}

// Automate a Go programming task with a LLM
// The input is a prompt and an optional starting point.
// The output is the generated / modified source code in a containerized dev environment
func (m *Demo) GoProgrammer(ctx context.Context,
	// A starting point for the coder's workspace.
	// Defaults to an empty directory
	// +optional
	start *dagger.Directory,
	// A description of the Go programming task to perform
	assignment string,
) (*dagger.Container, error) {
	progress := dag.Github().NewProgressReport(assignment, m.Token, m.Repo, m.Issue)
	// Send the initial progress report (we will update it later)
	progress = dag.Llm().
		WithGithubProgressReport(progress).
		WithPromptVar("assignment", assignment).
		WithPromptFile(dag.CurrentModule().Source().File("prompts/coder.txt")).
		GithubProgressReport()
	if err := progress.Publish(ctx); err != nil {
		return nil, err
	}
	// Initialize a Go-specific workspace
	workspace := dag.Workspace(dagger.WorkspaceOpts{
		Start:   start,
		Checker: dag.Go(dag.Directory()).Base().WithDefaultArgs([]string{"go", "build", "./..."}),
	})
	// Implement (single pass, no loop)
	coder := dag.
		Llm().
		WithWorkspace(workspace).
		WithPromptVar("assignment", assignment).
		WithPromptFile(dag.CurrentModule().Source().File("prompts/reporter-start.txt"))
	// Save the modified workspace
	workspace = coder.Workspace()
	// Inspect t he workspace history, publish it as tasks in the progress report
	// FIXME: do this on-the-fly within the devloop
	history, err := workspace.History(ctx)
	if err != nil {
		return nil, err
	}
	for i, change := range history {
		progress = progress.StartTask(fmt.Sprintf("dev-%d", i+1), change, "✅")
	}
	// Get the result in diff format, and add it to the progress report
	result, err := coder.Workspace().Diff(ctx)
	if err != nil {
		return nil, err
	}
	progress = progress.
		AppendSummary(fmt.Sprintf("\n### Result\n\n```\n%s\n```\n", result))
	if err := progress.Publish(ctx); err != nil {
		return nil, err
	}
	// Show the result in an interactive terminal, for convenience
	return dag.Container().From("golang").WithDirectory(".", coder.Workspace().Dir()), nil
}

// Automate a Go programming task with a LLM and create a Pull request
// The input is a prompt and an optional starting point.
// The output is the generated / modified source code in a containerized dev environment
func (m *Demo) GoProgrammerPr(ctx context.Context,
	// A starting point for the coder's workspace.
	// Defaults to an empty directory
	// +optional
	start *dagger.Directory,
	// A description of the Go programming task to perform
	// +optional
	assignment string,
	// Fork repo name
	// +optional
	forkName string,
	// Fork the upstream Repo
	// +optional
	fork bool,
) (string, error) {
	// Get assignment from issue if not provided
	if assignment == "" {
		progress := dag.Github().NewProgressReport(assignment, m.Token, m.Repo, m.Issue)
		issueBody, err := progress.Issue().Body(ctx)
		if err != nil {
			return "", err
		}
		assignment = issueBody
	}
	work, err := m.GoProgrammer(ctx, start, assignment)
	if err != nil {
		return "", err
	}
	changes := work.Directory(".")
	// Determine PR title
	title, err := dag.Llm().
		WithPrompt(fmt.Sprintf(
			`You will be given an input.
Summarize it to a short title, suitable as the title of a pull request for the assignment. Be extremely brief.
<input>
%s
</input>
`, assignment)).LastReply(ctx)
	if err != nil {
		return "", err
	}
	title = strings.Trim(title, "\"")

	// Use assignment for PR body
	body := "Assignment: " + assignment

	// Determine branch name
	branch, err := dag.Llm().
		WithPrompt(fmt.Sprintf(
			`You will be given an input.
Come up with a short suitable git branch name for a change set solving the assignment. The branch name should be no more than 20 alphanumeric characters.
<input>
%s
</input>
`, assignment)).LastReply(ctx)
	if err != nil {
		return "", err
	}

	// Lookup git remote
	remote, err := dag.FeatureBranch().
		WithGithubToken(m.Token).
		WithChanges(changes).
		GetRemoteURL(ctx, "origin")
	if err != nil {
		return "", err
	}

	fbCreateOpts := dagger.FeatureBranchCreateOpts{}
	if forkName != "" {
		fbCreateOpts.ForkName = forkName
	}
	if fork {
		fbCreateOpts.Fork = fork
	}

	return dag.FeatureBranch().
		WithGithubToken(m.Token).
		Create(remote, branch, fbCreateOpts).
		WithChanges(changes.WithoutDirectory(".git")).
		PullRequest(ctx, title, body)
}

func parseGithubUrl(url string) (string, string, bool) {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return "", "", false
	}
	// Remove github.com prefix if present
	if parts[0] == "github.com" {
		parts = parts[1:]
	}
	return parts[0], parts[1], true
}
