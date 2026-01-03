@echo off
echo Generating protobuf files...

REM Python protobuf
echo Generating Python protobuf...
cd docreader
python -m grpc_tools.protoc -I./proto --python_out=./proto --grpc_python_out=./proto proto/docreader.proto

REM Go protobuf
echo Generating Go protobuf...
cd ..\backend
if not exist pkg\proto mkdir pkg\proto
protoc --go_out=./pkg/proto --go-grpc_out=./pkg/proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative -I=../docreader/proto ../docreader/proto/docreader.proto

cd ..
echo Protobuf files generated successfully!
