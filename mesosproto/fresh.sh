#!/bin/bash

# See:
# https://github.com/golang/protobuf
# https://github.com/gogo/protobuf

#
# ---> Install protoc
#
# yum -y install protobuf protobuf-compiler
# ls /usr/bin/protoc

#
# ---> Install protoc-gen-gogo:
#
# go get github.com/gogo/protobuf/proto
# go get github.com/gogo/protobuf/jsonpb
# go get github.com/gogo/protobuf/protoc-gen-gogo
# go get github.com/gogo/protobuf/gogoproto
#
# ---> Using protoc-gen-gogo
#
protoc --proto_path=. --gogo_out=. *.proto

#
# OR
#

#
# ---> Install protoc-gen-go (Not Required Any More?)
#
# go get github.com/golang/protobuf/protoc-gen-go
#
# ---> Using protoc-gen-go
#
# protoc --proto_path=. --go_out=. *.proto
