#!/bin/sh

set -xe

which dagger 2>/dev/null || {
    echo >&2 "Dagger not installed. Follow installation instructions at https://docs.dagger.io/install"
    exit 1
}
if [ -z "${OPENAI_API_KEY}" ] && { [ ! -f .env ] || ! grep -q "^OPENAI_API_KEY=" .env; }; then
    printf "Enter OpenAI API key (https://platform.openai.com/api-keys), or press enter to skip: "
    read OPENAI_API_KEY

    if [ ! -z "${OPENAI_API_KEY}" ]; then
        cat <<EOF > .env
OPENAI_API_KEY=${OPENAI_API_KEY}
EOF
    fi
fi

if [ -z "${ANTHROPIC_API_KEY}" ] && { [ ! -f .env ] || ! grep -q "^ANTHROPIC_API_KEY=" .env; }; then
    printf "Enter Anthropic API key (https://console.anthropic.com/account/keys), or press enter to skip: "
    read ANTHROPIC_API_KEY

    if [ ! -z "${ANTHROPIC_API_KEY}" ]; then
        cat <<EOF >> .env
ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
EOF
    fi
fi

echo "--- Installing CLI: $HOME/bin/dagger-llm"
dagger shell <<EOF
./dagger-llm | cli current | export $HOME/bin/dagger-llm
EOF

echo "--- Running Engine"
echo "--- To connect: _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:1234 ~/bin/dagger-llm shell"
dagger shell <<EOF
./dagger-llm | engine | up
EOF
