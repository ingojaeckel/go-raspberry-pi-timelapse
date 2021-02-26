#!/bin/sh
set -e -x
GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`
VERSION="${GIT_COMMIT} / ${BUILT_AT}" npm run build
