#!/usr/bin/env bash
# -*- coding: UTF-8 -*-
echo "Kubernetes Community release script"

# Project name as it appear community-operator-prod, e.g. "node-healthcheck-operator"
PROJECT=$1
# Tagged version, e.g. "0.3.0" (no "v" prefix)
VERSION=$2

echo "Creating a new bundle"
rm -r bundle
make bundle

echo "Cloning the community repo (my fork for testing)"
set -xe
git clone --depth 1 https://github.com/clobrano/community-operators-prod

echo "Copy bundle into the community repo"
mkdir community-operators-prod/operators/${PROJECT}/${VERSION}
cp -r bundle community-operators-prod/operators/${PROJECT}/${VERSION}

echo "Commit and push remote"
cd community-operators-prod
git switch -c add-${PROJECT}-${VERSION}
git add operators/${PROJECT}/${VERSION}
git commit -m "Add ${PROJECT} v${VERSION}"
git push origin add-${PROJECT}-${VERSION}

#echo "Clean up"
#cd ..
#rm -r community-operators-prod
