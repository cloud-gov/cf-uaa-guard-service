#!/bin/bash

set -e

go get -v github.com/Masterminds/glide

pushd broker
  glide install
popd

pushd proxy
  glide install
popd
