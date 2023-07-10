#!/bin/bash

docker login quay.io

docker build \
 --build-arg PROGRAM_NAME=xdp_pass \
 --build-arg SECTION_NAME=pass \
 --build-arg PROGRAM_TYPE=xdp \
 --build-arg BYTECODE_FILENAME=xdp_pass.bpf.o \
 -f ../../../packaging/container-deployment/Containerfile.bytecode \
 ./.output -t quay.io/bpfd-bytecode/xdp_pass:latest

docker push quay.io/bpfd-bytecode/xdp_pass

docker build \
 --build-arg PROGRAM_NAME=tc_pass \
 --build-arg SECTION_NAME=pass \
 --build-arg PROGRAM_TYPE=tc \
 --build-arg BYTECODE_FILENAME=tc_pass.bpf.o \
 -f ../../../packaging/container-deployment/Containerfile.bytecode \
 ./.output -t quay.io/bpfd-bytecode/tc_pass:latest

docker push quay.io/bpfd-bytecode/tc_pass

docker build \
 --build-arg PROGRAM_NAME=tracepoint \
 --build-arg SECTION_NAME=sys_enter_openat \
 --build-arg PROGRAM_TYPE=tracepoint \
 --build-arg BYTECODE_FILENAME=tp_openat.bpf.o \
 -f ../../../packaging/container-deployment//Containerfile.bytecode \
 ./.output -t quay.io/bpfd-bytecode/tracepoint:latest

docker push quay.io/bpfd-bytecode/tracepoint

docker build \
 --build-arg PROGRAM_NAME=mallocprobe \
 --build-arg SECTION_NAME=malloc \
 --build-arg PROGRAM_TYPE=uprobe \
 --build-arg BYTECODE_FILENAME=mallocprobe.bpf.o \
 -f ../../../packaging/container-deployment//Containerfile.bytecode \
 ./.output -t quay.io/bpfd-bytecode/mallocprobe:latest

docker push quay.io/bpfd-bytecode/mallocprobe

