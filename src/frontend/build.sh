#!/bin/sh
set -e

cd "$(dirname "$0")"

# Use the Node version from .nvmrc. If the active Node already matches (e.g. in
# CI, where actions/setup-node installs it outside nvm), leave it alone: nvm
# would report the version as "not installed" and fail the build.
WANT="v$(cat .nvmrc)"
HAVE="$(node --version 2>/dev/null || true)"
if [ "$HAVE" != "$WANT" ]; then
    export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
    if [ -s "$NVM_DIR/nvm.sh" ]; then
        . "$NVM_DIR/nvm.sh"
        nvm use >/dev/null || echo "Warning: nvm could not switch to $WANT (run 'nvm install'); using ${HAVE:-system Node}"
    else
        echo "Warning: Node $WANT wanted, using ${HAVE:-system Node} (nvm not found)"
    fi
fi
echo "Node $(node --version)"

set -x
COMMIT=$(git rev-parse HEAD)
COMMIT_ABBREV=$(git rev-parse --short HEAD)
COMMIT_TIME=$(git log -1 --format=%cd)
CI=true \
    REACT_APP_GIT_SHA=${COMMIT} \
    REACT_APP_GIT_SHA_ABBREV=${COMMIT_ABBREV} \
    REACT_APP_COMMIT_TIME=${COMMIT_TIME} \
    npm run build
