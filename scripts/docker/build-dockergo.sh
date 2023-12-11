#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
function buildImage() {
  rm -rf vm-engine
  git clone https://git.code.tencent.com/ChainMaker/vm-engine.git
  cd vm-engine
  git checkout v2.3.2_qc
  make build-image
  cd ..
  rm -rf vm-engine
}


dockerGoImage=$( docker images | grep "^chainmakerofficial/chainmaker-vm-engine" | grep "v2.3.2\s" )
echo "image:" $dockerGoImage
if [[ -n $dockerGoImage ]] ;then
    echo "docker go image exist, don't need build again"
else
    echo "build new docker go image"
    buildImage
fi