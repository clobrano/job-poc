#!/usr/bin/env bash
# -*- coding: UTF-8 -*-
set -eu

# Tagged version, e.g. "0.3.0" (no "v" prefix)
VERSION=$1
# Project name as it appear community-operator-prod, e.g. "node-healthcheck-operator"
PROJECT=${2:-${PWD##*/}}
# community-operators-prod fork username
USERNAME=${3:-${USER}}
OCP_VERSION=v4.11

echo "Creating Red Hat Community release"
echo "Project: ${PROJECT}"
echo "Version: ${VERSION}"
echo "OCP Version: ${OCP_VERSION}"
echo "Fork repository where to push the change: https://github.com/${USERNAME}/community-operators-prod"

read -p "Press ENTER to continue CTLR-C to abort"

echo "Creating a new bundle"
rm -r bundle
export VERSION=${VERSION}
make bundle-build bundle-community

cat <<EOF >> bundle/metadata/annotations.yaml

  # Annotations for OCP
  com.redhat.openshift.versions: "${OCP_VERSION}"
EOF

echo "Cloning the community repo (my fork for testing)"
set -xe
git clone --depth 1 https://github.com/${USERNAME}/community-operators-prod

echo "Copy bundle into the community repo"
mkdir community-operators-prod/operators/${PROJECT}/${VERSION}
cp -r bundle/* community-operators-prod/operators/${PROJECT}/${VERSION}

echo "Commit and push remote"
cd ./community-operators-prod
git switch -c add-${PROJECT}-${VERSION}
git add operators/${PROJECT}/${VERSION}
git commit -sm "Add ${PROJECT} v${VERSION}"
git push origin add-${PROJECT}-${VERSION}
git log -1

cd ..
echo "Cleaning up"
sudo rm -r community-operators-prod
xdg-open https://github.com/${USERNAME}/community-operators-prod
