#!/usr/bin/env bash

echo STABLE_VERSION "$VERSION"
echo VERSION_TIMESTAMP "$(date +%s)"

# Unstable Buildkite env vars
echo "BUILDKITE $BUILDKITE"
echo "BUILDKITE_COMMIT $BUILDKITE_COMMIT"
echo "BUILDKITE_BRANCH $BUILDKITE_BRANCH"
echo "BUILDKITE_PULL_REQUEST_REPO $BUILDKITE_PULL_REQUEST_REPO"
echo "BUILDKITE_PULL_REQUEST $BUILDKITE_PULL_REQUEST"
