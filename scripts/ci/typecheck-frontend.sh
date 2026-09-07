#!/usr/bin/env bash
#
# GoPCA Suite
#
# Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
#
# This file is part of GoPCA Suite.
#
# See LICENSE for the full license terms.
#
# Type-checks all three frontends: the shared UI components, GoPCA Desktop and
# GoCSV.
#
# Why this is a script rather than a handful of workflow steps. The type error
# that broke `make pca-dev` in #885 passed a completely green pull request, and
# the first attempt to catch it in CI then passed on a developer machine and
# failed on a runner. Both failures have one cause: the check ran against a tree
# in a different state from the one it was meant to verify. A single script,
# invoked identically by `make typecheck` and by the workflow, removes that
# difference -- what a developer runs before pushing is line for line what CI
# runs afterwards.
#
# Type-checking is only meaningful against real bindings. `tsc` either resolves
# the `wailsjs` imports to actual Go-derived declarations, or it reports TS2307
# for every one of them followed by a cascade of TS7006 on the values that are
# now untyped. Stubbing those imports would make the check pass without
# checking the thing that matters, so this script generates them.

set -euo pipefail

cd "$(dirname "$0")/../.."
ROOT="$(pwd)"

# Locate the Wails CLI, following the same convention as build-desktop.sh.
if ! command -v wails &> /dev/null; then
    if [ -x "$(go env GOPATH)/bin/wails" ]; then
        export PATH="$PATH:$(go env GOPATH)/bin"
    else
        echo "ERROR: Wails CLI is not installed, and the frontends cannot be"
        echo "type-checked without the Go bindings it generates."
        echo "Install with: go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0"
        exit 1
    fi
fi

# `wails generate module` runs `go build -tags bindings` on the app and executes
# the result, and both main.go files carry `//go:embed all:frontend/dist`. That
# directory is a build artifact absent from a fresh checkout, so the build fails
# with "pattern all:frontend/dist: no matching files found" before it can emit
# anything. An empty placeholder satisfies the embed -- the `all:` prefix makes
# the dotfile a match -- and the bindings derive from the Go source, not from
# that directory's contents. An existing dist is left alone.
#
# Building the real dist here is not an option: `npm run build` is
# `tsc && vite build`, and tsc is precisely what needs the bindings.
generate_bindings() {
    local app="$1"
    if [ -z "$(ls -A "$ROOT/cmd/$app/frontend/dist" 2>/dev/null)" ]; then
        mkdir -p "$ROOT/cmd/$app/frontend/dist"
        touch "$ROOT/cmd/$app/frontend/dist/.gitkeep"
    fi
    (cd "$ROOT/cmd/$app" && wails generate module)
}

# The apps import their types from the built shared package, not from its
# source, so a missing or stale dist produces type errors that have nothing to
# do with the code under test. Rebuilding takes a few seconds; a check that can
# fail for a reason unrelated to the code is worth more than those seconds.
echo "==> Building shared UI components"
npm run build-ui

# Both apps' frontend/wailsjs/ are gitignored and generated, so neither can go
# stale against the Go it describes.
echo "==> Generating GoPCA Desktop Wails bindings"
generate_bindings gopca-desktop

echo "==> Generating GoCSV Wails bindings"
generate_bindings gocsv

# --no-install pins tsc to the version npm installed. Plain `npx tsc` silently
# fetches a TypeScript of its own choosing when one is missing from
# node_modules, and a check whose version can change underneath you is not a
# check.
echo "==> Type-checking shared UI components"
npx --no-install tsc --noEmit -p packages/ui-components/tsconfig.json

echo "==> Type-checking GoPCA Desktop frontend"
npx --no-install tsc --noEmit -p cmd/gopca-desktop/frontend/tsconfig.json

echo "==> Type-checking GoCSV frontend"
npx --no-install tsc --noEmit -p cmd/gocsv/frontend/tsconfig.json

echo ""
echo "All frontends type-check cleanly."
