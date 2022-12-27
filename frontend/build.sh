#!/bin/sh
set -e -x
COMMIT=`git rev-parse HEAD`
COMMIT_ABBREV=`git rev-parse --short HEAD`
COMMIT_TIME=`git log -1 --format=%cd`
CI=true \
    REACT_APP_GIT_SHA=${COMMIT} \
    REACT_APP_GIT_SHA_ABBREV=${COMMIT_ABBREV} \
    REACT_APP_COMMIT_TIME=${COMMIT_TIME} \
    npm run build
