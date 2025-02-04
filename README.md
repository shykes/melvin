# Melvin: a toy open-source programming agent

*WARNING: unfinished, experimental software*

This project is for educational purposes and should not be used for actual production-grade programming tasks.
There are several great "programming agent" products on the market, including [Devin](https://devin.ai), [Lovable](https://lovable.dev), [Replit Agent](https://replit.com), [V0](https://v0.dev) and many more.

## Internals

Melvin is distributed as a module for [Dagger](https://dagger.io), the programmable dag engine.
We use Dagger to break down Melvin into micro-agents, which can be run and tested independently,
or assembled into a broader software development workflow.

Melvin's core modules are written in Go, but it can be extended by any language,
since Dagger modules are cross-language.

You can run Melvin locally from the command-line; programmatically from your own software (Dagger has SDKs for Python, Go and Typescript);
or as a MCP tool server to be driven by another agent (coming soon).

## Getting started

1. Execute the `setup.sh` script to build and run the development version of Dagger, that Melvin requires

```console
./setup.sh
```

2. In another terminal, run a dagger interactive shell configured to use the dev build of dagger:

```console
_EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:1234 ~/bin/dagger-llm shell
```

3. You now have an interactive session with the Melvin module loaded. All the components of Melvin are available for you to run and compose.

Try:

```console
llm |
with-melvin-workspace --checker=$(container | from golang | with-default-args go build .) |
with-prompt "write a go program that computes the first 100 decimals of Pi, and prints them to the screen" |
melvin-workspace |
dir |
terminal
```
