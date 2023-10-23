#!/usr/bin/env bash

GENERATOR_DIR=$PWD

VER="21.1"
ARCH="aarch"
unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     OS=linux;;
    Darwin*)    OS=osx;;
    *)          OS="xxx"
esac

echo ${GENERATOR_DIR} ${OS} $VER $ARCH

if [ ! -d ${GENERATOR_DIR}/bin ]; then
mkdir ${GENERATOR_DIR}/bin
fi
if [ ! -d ${GENERATOR_DIR}/dist/protoc ]; then
mkdir -p ${GENERATOR_DIR}/dist/protoc
fi

if [ ! -f ${GENERATOR_DIR}/dist/protoc/protoc.zip ]; then
  echo "download protoc https://github.com/protocolbuffers/protobuf/releases/download/v$VER/protoc-$VER-$OS-${ARCH}_64.zip"
  curl https://github.com/protocolbuffers/protobuf/releases/download/v$VER/protoc-$VER-$OS-${ARCH}_64.zip -o $GENERATOR_DIR/dist/protoc/protoc.zip -L
fi

(cd ./dist/protoc && unzip -o protoc.zip)
cp -f ${GENERATOR_DIR}/dist/protoc/bin/protoc ${GENERATOR_DIR}/bin/protoc
echo "installed "`${GENERATOR_DIR}/bin/protoc --version`


pwd
for genpkg in `go list -f '{{ join .Imports "\n" }}' deps.go`
do
    echo "building $genpkg..."
    go get ${genpkg}
    go build -mod=readonly -o ${GENERATOR_DIR}/bin/`basename $genpkg` -trimpath ${genpkg}
    echo "installed $genpkg"
done

