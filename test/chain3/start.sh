#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

export LD_LIBRARY_PATH=$(dirname $PWD)/:$LD_LIBRARY_PATH
export PATH=$(dirname $PWD)/prebuilt/linux:$(dirname $PWD)/prebuilt/win64:$PATH
export WASMER_BACKTRACE=1
cp -rf ../../bin ./
export CMC=$PWD/bin
echo "cmc path:" $CMC

cd bin


pid=`ps -ef | grep chainmaker | grep "\-c ../config/node1/chainmaker.yml ci-chain3" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid} ];then
    nohup ./chainmaker start -c ../config/node1/chainmaker.yml ci-chain3 > panic1.log 2>&1 &
    echo "node1 chainmaker is starting, pls check log..."
else
    echo "node1 chainmaker is already started"
fi

pid2=`ps -ef | grep chainmaker | grep "\-c ../config/node2/chainmaker.yml ci-chain3" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid2} ];then
    nohup ./chainmaker start -c ../config/node2/chainmaker.yml ci-chain3 > panic2.log 2>&1 &
    echo "node2 chainmaker is starting, pls check log..."
else
    echo "node2 chainmaker is already started"
fi



pid3=`ps -ef | grep chainmaker | grep "\-c ../config/node3/chainmaker.yml ci-chain3" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid3} ];then
    nohup ./chainmaker start -c ../config/node3/chainmaker.yml ci-chain3 > panic3.log 2>&1 &
    echo "node3 chainmaker is starting, pls check log..."
else
    echo "node3 chainmaker is already started"
fi


pid4=`ps -ef | grep chainmaker | grep "\-c ../config/node4/chainmaker.yml ci-chain3" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid4} ];then
    nohup ./chainmaker start -c ../config/node4/chainmaker.yml ci-chain3 > panic4.log 2>&1 &
    echo "node4 chainmaker is starting, pls check log..."
else
    echo "node4 chainmaker is already started"
fi

# nohup ./chainmaker start -c ../config/node5/chainmaker.yml ci-chain3 > panic.log &

sleep 4
ps -ef|grep chainmaker | grep "ci-chain3"
