#!/bin/bash
#
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

CONTAINER_NAME=chainmaker-vm-go

docker ps

echo

read -r -p "input container name to stop(default 'chainmaker-vm-go'): " tmp
if  [ -n "$tmp" ] ;then
  CONTAINER_NAME=$tmp
else
  echo "container name use default: 'chainmaker-vm-go'"
fi

echo "stop docker vm container"

docker stop "$CONTAINER_NAME"
