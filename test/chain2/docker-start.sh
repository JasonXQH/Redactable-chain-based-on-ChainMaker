#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
VM_GO_IMAGE_NAME="chainmakerofficial/chainmaker-vm-engine:v2.3.2"

set -x
function xsed() {
    system=$(uname)

    if [ "${system}" = "Linux" ]; then
        sed -i "$@"
    else
        sed -i '' "$@"
    fi
}
# if enable docker vm service and use unix domain socket, run a vm docker container
#参数： 1 orgId，2 mountPath，3 logPath，4 节点端口，5，$contract_engine_portDocker映射端口
function start_vm_go() {
  container_name=VM-GO-wx-org$1.ci-chain2
  #check container exists
  exist=$(docker ps -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already RUNNING, please stop it first."
    exit 1
  fi

  exist=$(docker ps -a -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already exists(STOPPED)"
    if [ "$FORCE_CLEAN" == "true" ]; then
      docker rm $container_name > /dev/null
    else
      need_rm="yes"
      read -r -p "remove it and start a new container, default: yes (y|n): " need_rm
      if [ "$need_rm" == "no" ] || [ "$need_rm" == "n" ]; then
        exit 0
      else
        docker rm $container_name > /dev/null
      fi
    fi
  fi

  # concat mount_path and log_path for container to mount
  mount_path=$2
  log_path=$3
  if [[ "${mount_path:0:1}" != "/" ]];then
    mount_path=$(pwd)/$mount_path
  fi
  if [[ "${log_path:0:1}" != "/" ]];then
    log_path=$(pwd)/$log_path
  fi

  mkdir -p "$mount_path"
  mkdir -p "$log_path"

  EXPOSE_PORT=$5

  docker run -itd \
  --net=host \
  -v "$mount_path":/mount \
  -v "$log_path":/log \
  -e CHAIN_RPC_PORT="$5" \
  -e SANDBOX_RPC_PORT="$4" \
  --name $container_name \
  --privileged $VM_GO_IMAGE_NAME \
   > /dev/null


  retval="$?"
  if [ $retval -ne 0 ]; then
    echo "Fail to run docker vm."
    exit 1
  fi

  echo "start docker vm service container succeed: $container_name"
  xsed "s%enable: false%enable: true%g" ./config/wx-org$1.chainmaker.org/chainmaker.yml

}

echo "clean docker go"
docker rm -f  `docker ps -aq -f name=ci-chain2`
echo "start docker go"
start_vm_go 1 "./data/wx-org1.chainmaker.org/go" "./log/wx-org1.chainmaker.org/go" 32351 22351
start_vm_go 2 "./data/wx-org2.chainmaker.org/go" "./log/wx-org2.chainmaker.org/go" 32352 22352
start_vm_go 3 "./data/wx-org3.chainmaker.org/go" "./log/wx-org3.chainmaker.org/go" 32353 22353
start_vm_go 4 "./data/wx-org4.chainmaker.org/go" "./log/wx-org4.chainmaker.org/go" 32354 22354
sleep 3
docker ps | grep "ci-chain2"