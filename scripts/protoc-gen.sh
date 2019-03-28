#!/bin/bash
protoc -I${PWD}/../api/protobuf/v1 -I${PWD}/../third_party --go_out=plugins=grpc:${PWD}/../pkg/api/v1 ${PWD}/../api/protobuf/v1/vpn.proto
protoc -I${PWD}/../api/protobuf/v1 -I${PWD}/../third_party --grpc-gateway_out=logtostderr=true:${PWD}/../pkg/api/v1 ${PWD}/../api/protobuf/v1/vpn.proto
protoc -I${PWD}/../api/protobuf/v1 -I${PWD}/../third_party --swagger_out=logtostderr=true:${PWD}/../api/swagger/v1 ${PWD}/../api/protobuf/v1/vpn.proto