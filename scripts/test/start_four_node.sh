#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
## deploy ChainMaker and test

CURRENT_PATH=$(pwd)
PROJECT_PATH=$(dirname $(dirname "${CURRENT_PATH}"))
#echo "PROJECT_PATH $PROJECT_PATH"

cd $PROJECT_PATH
if [ -e bin/chainmaker ]; then
    echo "skip make, chainmaker binary already exist"
else
    make
fi

cd bin
nohup ./chainmaker start -c ../config/wx-org1/chainmaker.yml start_four_node > panic1.log 2>&1 &
nohup ./chainmaker start -c ../config/wx-org2/chainmaker.yml start_four_node > panic2.log 2>&1 &
nohup ./chainmaker start -c ../config/wx-org3/chainmaker.yml start_four_node > panic3.log 2>&1 &
nohup ./chainmaker start -c ../config/wx-org4/chainmaker.yml start_four_node > panic4.log 2>&1 &
echo "start chainmaker..."