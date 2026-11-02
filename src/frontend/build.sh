#!/bin/sh
set -e -x

# Load nvm and use the version specified in .nvmrc
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [ -s "$NVM_DIR/nvm.sh" ]; then
    . "$NVM_DIR/nvm.sh"
    nvm use
else
    echo "Warning: nvm not found, using system Node/npm"
fi

COMMIT=`git rev-parse HEAD`
COMMIT_ABBREV=`git rev-parse --short HEAD`
COMMIT_TIME=`git log -1 --format=%cd`
CI=true \
    REACT_APP_GIT_SHA=${COMMIT} \
    REACT_APP_GIT_SHA_ABBREV=${COMMIT_ABBREV} \
    REACT_APP_COMMIT_TIME=${COMMIT_TIME} \
    npm run build
