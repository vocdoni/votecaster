#!/bin/bash
#
protoc --go_out=. --go_opt=paths=source_relative hubmessage.proto username_proof.proto