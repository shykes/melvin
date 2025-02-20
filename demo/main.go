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
		WithPromptFile(dag.CurrentModule().Source().File("prompts/reporter-start.txt")).
		GithubProgressReport()
	if err := progress.Publish(ctx); err != nil {
		return nil, err
	}
	// Initialize a Go-specific workspace
	workspace := dag.Workspace(dagger.WorkspaceOpts{
		Start:    start,
		Checkers: []*dagger.WorkspaceChecker{dag.GoChecker().AsWorkspaceChecker()},
		OnSave:   []*dagger.WorkspaceNotifier{progress.AsWorkspaceNotifier()},
	})
	var (
		diff string
		err  error
	)
	// Implement & review loop
	for i := 1; ; i++ {
		workspace = dag.
			Llm().
			WithWorkspace(workspace).
			WithPromptVar("assignment", assignment).
			WithPromptFile(dag.CurrentModule().Source().File("prompts/coder.txt")).
			Workspace()
		diff, err = workspace.Diff(ctx)
		if err != nil {
			return nil, err
		}
		reviewed := dag.Reviewer().AddReview(assignment, workspace.Dir(), diff)
		score, err := dag.Reviewer().Score(ctx, reviewed)
		if err != nil {
			return nil, err
		}
		summary, err := reviewed.File(".review/summary").Contents(ctx)
		if err != nil {
			return nil, err
		}
		status := "❗"
		if score >= 7 {
			status = "✅"
		}
		progress = progress.
			StartTask(
				fmt.Sprintf("review-%d", i),
				fmt.Sprintf("Code review #%d", i),
				fmt.Sprintf("%s %d/10: %s", status, score, summary),
			)
		if err := progress.Publish(ctx); err != nil {
			return nil, err
		}
		if score >= 7 {
			break
		}
		workspace = workspace.CopyDir(".review", reviewed.Directory(".review"))
	}
	progress = progress.
		AppendSummary(fmt.Sprintf("\n### Result\n\n```\n%s\n```\n", diff))
	if err := progress.Publish(ctx); err != nil {
		return nil, err
	}
	// Show the result in an interactive terminal, for convenience
	return dag.Container().From("golang").WithDirectory(".", workspace.Dir()), nil
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
		WithPromptVar("input", assignment).
		WithPrompt(
			`You will be given an input.
Summarize it to a short title, suitable as the title of a pull request for the assignment. Be extremely brief.

<input>
$input
</input>
`).LastReply(ctx)
	if err != nil {
		return "", err
	}
	title = strings.Trim(title, "\"")

	// Use assignment for PR body
	body := "Assignment: " + assignment

	// Determine branch name
	branch, err := dag.Llm().
		WithPromptVar("input", assignment).
		WithPrompt(
			`You will be given an input.
Come up with a short suitable git branch name for a change set solving the assignment.
The branch name should be no more than 20 alphanumeric characters.
<input>
$input
</input>`).
		LastReply(ctx)
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
