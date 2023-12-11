#!/bin/bash
#
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

LOG_PATH=$(pwd)/log
MOUNT_PATH=$(pwd)/docker-go
LOG_LEVEL=INFO
EXPOSE_PORT=22351
RUNTIME_PORT=32351
CONTAINER_NAME=chainmaker-vm-go
IMAGE_NAME="chainmakerofficial/chainmaker-vm-engine:v2.3.2"


read -r -p "input path to cache contract files(must be absolute path, default:'./docker-go'): " tmp
if  [ -n "$tmp" ] ;then
  MOUNT_PATH=$tmp
fi

if  [ ! -d "$MOUNT_PATH" ];then
  read -r -p "contracts path does not exist, create it or not(y|n): " need_create
  if [ "$need_create" == "yes" ] || [ "$need_create" == "y" ]; then
    mkdir -p "$MOUNT_PATH"
    if [ $? -ne 0 ]; then
      echo "create contracts path failed. exit"
      exit 1
    fi
  else
    exit 1
  fi
fi


read -r -p "input log path(must be absolute path, default:'./log'): " tmp
if  [ -n "$tmp" ] ;then
  LOG_PATH=$tmp
fi

if  [ ! -d "$LOG_PATH" ];then
  read -r -p "log path does not exist, create it or not(y|n): " need_create
  if [ "$need_create" == "yes" ] || [ "$need_create" == "y" ]; then
    mkdir -p "$LOG_PATH"
    if [ $? -ne 0 ]; then
      echo "create log path failed. exit"
      exit 1
    fi
  else
    exit 1
  fi
fi

read -r -p "input log level(DEBUG|INFO(default)|WARN|ERROR): " tmp
if  [ -n "$tmp" ] ;then
  if  [ $tmp == "DEBUG" ] || [ $tmp == "INFO" ] || [ $tmp == "WARN" ] || [ $tmp == "ERROR" ];then
      LOG_LEVEL=$tmp
  else
    echo "unknown log level [" $tmp "], so use default"
  fi
fi

read -r -p "input expose port(default 22351): " tmp
if  [ -n "$tmp" ] ;then
  if [[ $tmp =~ ^[0-9]+$ ]] ;then
      EXPOSE_PORT=$tmp
  else
    echo "unknown expose port [" $tmp "], so use 22351"
  fi
fi

read -r -p "input runtime port(default 32351): " tmp
if  [ -n "$tmp" ] ;then
  if [[ $tmp =~ ^[0-9]+$ ]] ;then
      RUNTIME_PORT=$tmp
  else
    echo "unknown expose port [" $tmp "], so use 32351"
  fi
fi

read -r -p "input container name(default 'chainmaker-docker-vm'): " tmp
if  [ -n "$tmp" ] ;then
  CONTAINER_NAME=$tmp
else
  echo "container name use default: 'chainmaker-docker-vm'"
fi

read -r -p "input vm config file path(use default config(default)): " tmp
if  [ -n "$tmp" ] ;then
  DOCKER_VM_CONFIG_FILE=$tmp
  if [ ! -d $MOUNT_PATH/config  ];then
    mkdir $MOUNT_PATH/config
  fi
  cp $DOCKER_VM_CONFIG_FILE $MOUNT_PATH/config/
else
  echo "docker-vm config is nil, use default config"
fi

exist=$(docker ps -a -f name="$CONTAINER_NAME" --format '{{.Names}}')
if [ "$exist" ]; then
  echo "container is exist, please remove container first."
  exit 1
fi

echo "start docker vm container"

if [ "$DOCKER_VM_CONFIG_FILE" == "" ]; then
	docker run -itd --net=host \
		-v "$MOUNT_PATH":/mount \
		-v "$LOG_PATH":/log \
		-e CHAIN_RPC_PORT="$EXPOSE_PORT" \
		-e SANDBOX_RPC_PORT="$RUNTIME_PORT" \
		--name "$CONTAINER_NAME" \
		--privileged $IMAGE_NAME
else
	docker run -itd --net=host \
		-v "$MOUNT_PATH":/mount \
		-v "$LOG_PATH":/log \
		--name "$CONTAINER_NAME" \
		--privileged $IMAGE_NAME
fi

docker ps -a -f name="$CONTAINER_NAME"