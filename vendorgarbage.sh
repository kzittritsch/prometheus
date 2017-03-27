#!/bin/bash

if [[ -z $GOPATH ]]; then
    echo "you need a GOPATH"
    exit 1
fi

git remote -v | grep upstream
if [[ $? != 0 ]]; then
    echo "Adding upstream"
    git remote add upstream https://github.com/prometheus/prometheus.git
fi

git fetch upstream master
git merge upstream/master
if [[ $? != 0 ]]; then
    echo "You probably have merge conflicts"
    echo "Please fix"
    exit 1
fi

cp ./config/config.go $GOPATH/src/github.com/prometheus/prometheus/config/
cp ./discovery/discovery.go $GOPATH/src/github.com/prometheus/prometheus/discovery/
cp ./discovery/configgrid/configgrid.go $GOPATH/src/github.com/prometheus/prometheus/discovery/configgrid/
