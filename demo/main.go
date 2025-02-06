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
	// Parse the repo
	owner, repo, ok := parseGithubUrl(m.Repo)
	if !ok {
		return nil, fmt.Errorf("incorrect github repo address: %s", m.Repo)
	}
	// Extract a progress report title from the task description
	title, err := dag.Llm().
		WithPrompt(fmt.Sprintf(
			`You will be given an input.
Summarize it to a short title, suitable as the title of a status update document it to a status update.
<input>
%s
</input>
`, assignment)).LastReply(ctx)
	if err != nil {
		return nil, err
	}
	// Send the initial progress report
	report := dag.Github().
		NewProgressReport(assignment, m.Token, owner, repo, m.Issue).
		WriteTitle(title).
		StartTask("analyze-task", "Analyze task", "⏳").
		StartTask("analyze-code", "Analyze source code", "⏳").
		StartTask("implement", "Implement solution", "⏳").
		StartTask("test", "Test solution", "⏳").
		StartTask("pr", "Submit pull request", "⏳").
		StartTask("done", "User confirmation", "⏳")
	if err := report.Publish(ctx); err != nil {
		return nil, err
	}
	// Initialize the workspace, with a Go-specific checker
	workspace := dag.Workspace(dagger.WorkspaceOpts{
		Start:   start,
		Checker: dag.Go(dag.Directory()).Base().WithDefaultArgs([]string{"go", "build", "./..."}),
	})
	// Analyze the source code + assignment to produce a list of sub-tasks
	tasks, err := dag.Llm().
		WithWorkspace(workspace).
		WithPrompt(`
You are a Go programmer.
Given an assignment, and a workspace with code, analyze the contents of the workspace to make a plan for how to accomplish the assignment.
Produce a list of tasks that another programmer can easily follow.

Output tasks one per line, in standard markdown bullet list format. Output nothing else.

<assignment>` + assignment + `</assignment>`).
		LastReply(ctx)
	if err != nil {
		return nil, err
	}
	// Add the coding tasks to the progress report
	dag.Llm().WithGithubProgressReport(report).WithPrompt(fmt.Sprintf(
		`You are a status updater. You will be given an assignment, and a list of tasks needed to accomplish it.

		1) Send a summary explaining the assignment, and an overview of the planned tasks.
		2) Report each task as started, with the status "⏳"
		3) publish at the end

		Don't change the title please

		<assignment>
		%s
		</assignment>
		<tasks>
		%s
		</tasks>
		`, assignment, tasks)).
		LastReply(ctx)
	// Write the code! No loop for now
	coder := dag.
		Llm().
		WithWorkspace(workspace).
		WithPrompt(fmt.Sprintf(`
You are a Go programmer.
You will be given an assignment and a list of tasks to accomplish it.
Use your workspace to accomplish the tasks.
Check your work with the 'check' tool. Continue until your tasks are completed, and the check succeeds.

<assignment>
%s
</assignment>
<tasks>
%s
</tasks>
`, assignment, tasks))
	// Get the result. For now just a diff to
	result, err := coder.Workspace().Diff(ctx)
	if err != nil {
		return nil, err
	}
	if err := report.
		WriteSummary(fmt.Sprintf("%s\n\nHere is the result: ```\n%s\n```\n", report.Summary, result)).
		Publish(ctx); err != nil {
		return nil, err
	}
	return dag.Container().From("golang").WithDirectory(".", coder.Workspace().Dir()), nil
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
