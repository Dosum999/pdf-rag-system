#!/bin/bash

set -e

echo "Generating protobuf files..."

# Python protobuf
echo "Generating Python protobuf..."
cd docreader
python -m grpc_tools.protoc \
  -I./proto \
  --python_out=./proto \
  --grpc_python_out=./proto \
  proto/docreader.proto

# Go protobuf
echo "Generating Go protobuf..."
cd ../backend
mkdir -p pkg/proto
protoc \
  --go_out=./pkg/proto \
  --go-grpc_out=./pkg/proto \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  -I=../docreader/proto \
  ../docreader/proto/docreader.proto

echo "Protobuf files generated successfully!"
