// Here is a little Dagger module,
// which provides tools for a little "Devin-like"
// programming agent.
//
// With Dagger you don't have to build a giant monolith agent.
// You just compose tools. Then, sometimes, some of the tools
// prompt some of your objects. And those tools might be
// wrapped in another object that is itself prompted.
// So it's agents all the way down; like micro-services
// but for agents.
package main

import (
	"context"
	"dagger/melvin/internal/dagger"
	"errors"
	"strconv"
)

func New(
	// Git repository to work on
	repo string,
	// +optional¡
	githubToken *dagger.Secret,
) Melvin {
	return Melvin{
		Repo:        repo,
		GithubToken: githubToken,
	}
}

type Melvin struct {
	GithubToken *dagger.Secret
	Base        *dagger.Container
	Repo        string
}

func (m Melvin) Issue(issueId int) Issue {
	return Issue{
		Melvin:      m,
		IssueNumber: issueId,
		Workspace:   dag.Directory(),
	}
}

// A github issue on the project
type Issue struct {
	// The current Melvin
	Melvin Melvin
	// The number of the github issue to solve
	IssueNumber int
	// A private workspace for the current task
	Workspace *dagger.Directory
}

// Get the contents of the issue
func (solver Issue) Body(
	ctx context.Context,
	// +optional
	// Include all comments for the issue
	comments bool,
) (string, error) {
	result, err := solver.GithubCli(ctx, []string{"issue", "view", strconv.Itoa(solver.IssueNumber)})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", errors.New(result.Stderr)
	}
	return result.Stdout, nil
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Run a command with the github CLI (gh), pre-configured with the Melvin token and repo
func (solver Issue) GithubCli(ctx context.Context, args []string) (CommandResult, error) {
	ctr := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "add", "github-cli"}).
		WithSecretVariable("GITHUB_TOKEN", solver.Melvin.GithubToken).
		WithExec(
			append([]string{"gh", "-R", solver.Melvin.Repo}, args...),
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
	exitCode, err := ctr.ExitCode(ctx)
	if err != nil {
		return CommandResult{}, err
	}
	stdout, err := ctr.Stdout(ctx)
	if err != nil {
		return CommandResult{}, err
	}
	stderr, err := ctr.Stderr(ctx)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// Get the git repository for the Melvin
func (solver Issue) Repo() *dagger.GitRepository {
	return dag.Git(solver.Melvin.Repo)
}

// Send a comment on the github issue
func (solver Issue) SendComment(ctx context.Context,
	// The text of the comment to write. Please remember that your entire team, and possibly the entire internet, can read this!
	text string,
) error {
	result, err := solver.GithubCli(ctx, []string{"issue", "comment", "--body", text, strconv.Itoa(solver.IssueNumber)})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return errors.New(result.Stderr)
	}
	return nil
}

// Checkout the Melvin repository
func (solver Issue) Checkout(
	// The git branch or tag to checkout
	// Defaults to the head branch
	// +optional
	branchOrTag string,
	// The path in the workspace to checkout to
	// +optional
	// +default="./src"
	path string,
) Issue {
	repo := dag.Git(solver.Melvin.Repo)
	var ref *dagger.GitRef
	if branchOrTag != "" {
		ref = repo.Branch(branchOrTag)
	} else {
		ref = repo.Head()
	}
	solver.Workspace = solver.Workspace.WithDirectory(path, ref.Tree())
	return solver
}
