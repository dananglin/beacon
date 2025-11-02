#!/usr/bin/env bash

# This is a build script used as part of the deployment
# of Beacon to the Flow Platform.

set -o errexit
set -o nounset
set -o pipefail

export BEACON_APP_VERSION="${FLOW_SERVICE_VERSION}"

mage="go tool -modfile=./tools/tools.mod mage"
mkdir -p "${FLOW_SERVICE_BUILD_DIR}"
${mage} clean build
mv "__build/${BEACON_APP_NAME}" "${FLOW_SERVICE_BUILD_DIR}/${BEACON_APP_NAME}"
echo "Successfully moved the binary to ${FLOW_SERVICE_BUILD_DIR}/${BEACON_APP_NAME}."
