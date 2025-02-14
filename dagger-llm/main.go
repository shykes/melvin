// Melvin depends on a specific development version of Dagger.
// This module manages the building and running of that specific version - using Dagger :)
// This takes advantage of Dagger's ability to run Dagger-in-Dagger seamlessly.

package main

import (
	"dagger/dagger-llm/internal/dagger"
)

type DaggerLlm struct{}

// Return a directory with the Dagger CLI at ./bin/dagger-llm
// Export it to your home directory, or /usr/local, or the prefix
// of your choice.
func (m *DaggerLlm) Cli(
	platform dagger.Platform,
) *dagger.File {
	cli := dag.DaggerDev().Cli().Binary(dagger.DaggerDevCliBinaryOpts{
		Platform: platform,
	})
	// Rename the file
	return dag.Directory().
		WithFile("dagger-llm", cli).
		File("dagger-llm")
}

// Build the Dagger engine at the version needed by Melvin,
// and return it as a service endpoint ready to be started.
func (m *DaggerLlm) Engine() *dagger.Service {
	return dag.DaggerDev().Engine().Service("melvin-dev")
}
