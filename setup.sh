#!/bin/sh

set -xe

which dagger 2>/dev/null || {
    echo >&2 "Dagger not installed. Follow installation instructions at https://docs.dagger.io/install"
    exit 1
}
# Only prompt for OPENAI_API_KEY if .env doesn't exist or if OPENAPI_API_KEY not set in .env
if [ ! -f .env ] || ! grep -q "^OPENAI_API_KEY=" .env; then
    printf "Enter your OPENAI_API_KEY (plaintext, or reference uri: op:// vault:// env:// file://) "
    read OPENAI_API_KEY

    # Create the .env file with the provided key
    cat <<EOF >> .env
OPENAI_API_KEY=${OPENAI_API_KEY}
EOF
fi

echo "--- Installing CLI: $HOME/bin/dagger-llm"
dagger shell <<EOF
./dagger-llm | cli current | export $HOME/bin/dagger-llm
EOF

echo "--- Running Engine"
echo "--- To connect: _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:1234 ~/bin/dagger-llm shell"
echo "--- Run 'llm' as your first command (bug workaround)"
dagger shell <<EOF
./dagger-llm | engine | up
EOF
