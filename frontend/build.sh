#!/bin/sh
set -e -x
GIT_COMMIT=`git rev-parse --short HEAD`
CI=true REACT_APP_GIT_SHA=${GIT_COMMIT} npm run build
