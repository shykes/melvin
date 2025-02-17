# Melvin: a toy open-source programming agent

*WARNING: unfinished, experimental software*

This project is for educational purposes and should not be used for actual production-grade programming tasks.
There are several great "programming agent" products on the market, including [Devin](https://devin.ai), [Lovable](https://lovable.dev), [Replit Agent](https://replit.com), [V0](https://v0.dev) and many more.

## Architecture

Rather than a monolithic application, Melvin is a set of modular components that you can integrate into your application,
or use individually.

Each Melvin module has the following features:

- Runs in containers, for maximum portability and reproducibility
- Can be run from the command-line, or programmatically via an API
- Generated bindings for Go, Python, Typescript, PHP (experimental, Java (experimental), and Rust (experimental).
- End-to-end tracing of prompts, tool calls, and even low-level system operations. 100% of agent state changes are traced.
- Cross-language extensions. Add your own modules in any language.
- Platform-independent. No infrastructure lock-in! Runs on any hosting provider that can run containers.

These features are achieved by using the [Dagger Engine](https://dagger.io) as a runtime.
Dagger is open-source, and can be installed on any machine that can run Linux containers.
This includes Mac and Windows machines with a Docker-compatible tool installed.

Dagger is the only dependency for Melvin. No other tooling or programming environment is required:
the entire environment is containerized, for maximum portability.

## Modules

The following modules are currently implemented:

- [toy-workspace](./toy-workspace): a very, very simple development workspace for demo purposes
- [toy-programmer](./toy-programmer): a very, very simple programmer micro-agent for demo purposes
- [workspace](./workspace): a slightly more powerful workspace, with checkpoint and history features, and a configurable check function
- [reviewer](./reviewer): a code reviewer micro-agent
- [github](./github): a module for sending progress updates in a github issue
- [demo](./demo): a collection of demo functions tying the other modules together

## Community

We strongly recommend [joining our Community discord](https://discord.gg/KK3AfBP8Gw).
The Dagger community is very welcoming, and will be happy to answer your questions, discuss your use case ideas, and help you get started.

Do this now! It will make the rest of the experience more productive, and more fun.

## Run Melvin from the command-line

To run Melvin from the command-line:

- first [complete initial setup](#initial-setup).
- Load one of the Melvin modules from the dagger CLI, and run functions

For example:

```console
~/bin/dagger-llm shell <<EOF
./toy-programmer | go-program "develop a curl clone" | terminal
EOF
```

This loads the `./toy-programmer` module, calls the function `go-program` with a description of a
program to write, then runs the `terminal` function on the returned container.

You can also explore available functions interactively:

```console
~/bin/dagger-llm shell
```

Then use tab auto-completion to explore.


### Integrate Melvin in your application

You can embed Dagger modules into your application.
Supported languages are Python, Typescript, Go, Java, PHP - with more language support under development.

1. Initialize a Dagger module at the root of your application.
This doesn't need to be the root of your git repository - Dagger is monorepo-ready.

```console
~/bin/dagger-llm init
```

2. Install the modules you wish to load

For example, to install the Melvin toy-workspace module:

```console
~/bin/dagger-llm install github.com/shykes/melvin/toy-workspace
```

3. Install a generated client in your project

*TODO: this feature is not yet merged in a stable version of Dagger*

This will configure Dagger to generate client bindings for the language of your choice.

For example, if your project is a Python application:

```console
~/bin/dagger-llm client install python
```

4. Re-generate clients

*TODO: this feature is not yet merged in a stable version of Dagger*

Any time you need to re-generate your client, run:

```console
~/bin/dagger-llm client generate
```

## Initial Setup

### Option 1: Quickstart

To get started quickly:

1. Run `./setup.sh` in the root of the repo.
2. Run `~/bin/dagger-llm shell -c 'llm | with-prompt "llm, are you there?" | last-reply'
3. If you got a response from the LLM, congratulations! Setup is complete

### Option 2: Not-so-quick-start

If the quickstart script doesn't work, or if you want to understand each step of the setup, follow these instructions.

#### Overview

Melvin's only dependency is Dagger.

At the moment, it requires a *development version* of Dagger which adds native support for LLM prompting and tool calling.
Once this feature is merged (current target is 0.17), Melvin will support with a stable release of Dagger.

#### 1. Install Dagger

First, install the latest release of Dagger,
by following the [official intallation instructions](https://docs.dagger.io/install).

We will use stable Dagger to build and run the development version of Dagger.

#### 2. Build dagger-llm client

Run this command from the root of the `melvin` repository:

```console
dagger shell <<EOF
./dagger-llm | cli current | export $HOME/bin/dagger-llm
EOF
```

This builds the llm-enabled version of the Dagger CLI, and installs it at `~/bin/dagger-llm`.

The first build might take a few minutes. Subsequent builds will be much faster due to caching.


#### 3. Run the dagger-llm engine

Run this command from the root of the `melvin` repository:

```console
dagger shell <<EOF
./dagger-llm | engine | up
EOF
```

This builds the llm-enabled version of the Dagger engine, and runs it.
Leave it running for the duration of your use of Melvin.


#### 4. Configure Dagger CLI

To connect to the development engine when using the development CLI, you need to set an environment variable:

```console
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:1234
```

#### 5. Configure LLM integration

LLM integration requires connecting to an OpenAI endpoint (other LLM providers will be supported soon).

Write your OpenAI token to a `LLM_KEY` variable in the `.env` file in the current directory.
You can write the actual token plaintext, or a reference to the secret in 1password, Hashicorp vault, a local file,
or an env variable.

For example:

```console
$ cat .env
# Plaintext format
LLM_KEY=sk-pr....
```

```console
$ cat .env
# 1password reference format
LLM_KEY=op://Dev/askdjhsajkdhsajkdhaskjdsa/credential
```

```console
$ cat .env
# Hashicorp vault format
LLM_KEY=vault://kjsdfjksdhfjkdsfjkhdskjf
```

In any case, *make sure the file is in the current directory when you call the Dagger CLI*.

#### 6. Initialize LLM integration

After initial setup, you need to initialize the llm integration by running this command once:

```console
~/bin/dagger-llm shell -c llm | with-prompt "LLM, are you there?" | last-reply
```

After execution completes, you should see a response from the LLM printed.
If so, congratulations! You have complete setup. Time to use Melvin!
