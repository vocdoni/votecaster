#!/bin/bash

if ! which git >/dev/null; then
    echo "git is not installed, install it first"
fi

if ! which go >/dev/null; then
    echo "go is not installed, install it first"
fi

if ! which protoc >/dev/null; then
    echo "protoc is not installed, install it first"
fi

CURRENT_DIR=$(pwd)

download_protobufs() {
    echo "Downloading protobufs..."
    git clone  --no-checkout https://github.com/farcasterxyz/hub-monorepo.git $CURRENT_DIR/hub-monorepo
    cd $CURRENT_DIR/hub-monorepo
    git sparse-checkout set --cone
    git sparse-checkout set protobufs/schemas
    git checkout main
    find . -mindepth 1 -maxdepth 1 ! -name 'protobufs' -exec rm -rf {} +
}

install_protoc() {
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
}

generate_go() {
    mkdir -p $CURRENT_DIR/protobufs
    protoc -I=$CURRENT_DIR/hub-monorepo/protobufs/schemas \
        --proto_path=$CURRENT_DIR/hub-monorepo/protobufs/schemas \
        --go_out=$CURRENT_DIR/protobufs --go_opt=paths=source_relative \
        --go_opt=Mmessage.proto=golang-submitmessage/protobufs \
        --go_opt=Musername_proof.proto=golang-submitmessage/protobufs \
        $CURRENT_DIR/hub-monorepo/protobufs/schemas/message.proto $CURRENT_DIR/hub-monorepo/protobufs/schemas/username_proof.proto
}

clean() {
    rm -rf $CURRENT_DIR/hub-monorepo
}

download_protobufs
install_protoc
generate_go
clean

