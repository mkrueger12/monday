#!/bin/bash
set -e

echo "🔍 Detecting and installing dependencies..."

# Node.js
if [ -f package.json ]; then
    echo "📦 Found package.json - installing Node.js dependencies"
    npm install --silent
fi

# Python
if [ -f requirements.txt ]; then
    echo "🐍 Found requirements.txt - installing Python dependencies"
    pip install -r requirements.txt
elif [ -f Pipfile ]; then
    echo "🐍 Found Pipfile - installing Python dependencies with pipenv"
    pip install pipenv && pipenv install --deploy
elif [ -f pyproject.toml ]; then
    echo "🐍 Found pyproject.toml - installing Python dependencies"
    pip install .
fi

# Ruby
if [ -f Gemfile ]; then
    echo "💎 Found Gemfile - installing Ruby dependencies"
    bundle install --jobs 4 --retry 3
fi

# Go
if [ -f go.mod ]; then
    echo "🐹 Found go.mod - downloading Go dependencies"
    go mod download
fi

# Rust
if [ -f Cargo.toml ]; then
    echo "🦀 Found Cargo.toml - installing Rust dependencies"
    if command -v cargo >/dev/null 2>&1; then
        cargo fetch
    else
        echo "⚠️  Cargo not available, skipping Rust dependencies"
    fi
fi

echo "✅ Dependency installation complete"
