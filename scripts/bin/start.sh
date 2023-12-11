#!/bin/bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

export LD_LIBRARY_PATH=$(dirname $PWD)/lib:$LD_LIBRARY_PATH
export PATH=$(dirname $PWD)/lib:$PATH
export WASMER_BACKTRACE=1


function parse_yaml {
   local prefix=$2
   local s='[[:space:]]*' w='[a-zA-Z0-9_]*' fs=$(echo @|tr @ '\034')
   sed -ne "s|^\($s\):|\1|" \
        -e "s|^\($s\)\($w\)$s:$s[\"']\(.*\)[\"']$s\$|\1$fs\2$fs\3|p" \
        -e "s|^\($s\)\($w\)$s:$s\(.*\)$s\$|\1$fs\2$fs\3|p"  $1 |
   awk -F$fs '{
      indent = length($1)/2;
      vname[indent] = $2;
      for (i in vname) {if (i > indent) {delete vname[i]}}
      if (length($3) > 0) {
         vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
         printf("%s%s%s=\"%s\"\n", "'$prefix'",vn, $2, $3);
      }
   }'
}

config_file="../config/{org_id}/chainmaker.yml"
# config_file="../../config/wx-org1-solo/chainmaker.yml"
eval $(parse_yaml "$config_file" "chainmaker_")

VM_GO_IMAGE_NAME="chainmakerofficial/chainmaker-vm-engine:v2.3.2"
DOCKER_VM_IMAGE_NAME="chainmakerofficial/chainmaker-vm-docker-go:v2.3.1"
START_FULL_MODE=""

# read params
for i in "$@";do
    case "$i" in
        -y|-f|force)
            FORCE_CLEAN=true
            ;;
        alone)
            START_FULL_MODE=false
            ;;
        full)
            START_FULL_MODE=true
            ;;
        *)
            echo $"Usage: $0 {alone|full|-y|-f|force}"
            echo $"  full: start chainmaker node with the vm containers"
            echo $"  alone: start chainmaker node alone, WITHOUT vm containers"
            echo $"  -y|-f|force: force remove stopped vm containers and create new"
            exit
    esac
done


# if enable docker vm service and use unix domain socket, run a vm docker container
function start_vm_go() {
  container_name=VM-GO-{org_id}-{tagName}
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
  mount_path=$chainmaker_vm_go_data_mount_path
  log_path=$chainmaker_vm_go_log_mount_path
  if [[ "${mount_path:0:1}" != "/" ]];then
    mount_path=$(pwd)/$mount_path
  fi
  if [[ "${log_path:0:1}" != "/" ]];then
    log_path=$(pwd)/$log_path
  fi

  mkdir -p "$mount_path"
  mkdir -p "$log_path"

  protocol=$chainmaker_vm_go_protocol
  vm_go_log_level=$chainmaker_vm_go_log_level
  dockervm_config_path=$chainmaker_vm_go_dockervm_config_path
  runtime_server_host=$chainmaker_vm_go_runtime_server_host
  runtime_server_port=$chainmaker_vm_go_runtime_server_port
  contract_engine_port=$chainmaker_vm_go_contract_engine_port
  rpc_timeout=$chainmaker_vm_go_dial_timeout
  rpc_max_send_size=$chainmaker_vm_go_max_send_msg_size
  rpc_max_recv_size=$chainmaker_vm_go_max_recv_msg_size
  log_in_console=$chainmaker_vm_go_log_in_console
  max_concurrency=$chainmaker_vm_go_max_concurrency
  slow_disable=$chainmaker_vm_go_slow_disable
  slow_step_time=$chainmaker_vm_go_slow_step_time
  slow_tx_time=$chainmaker_vm_go_slow_tx_time
  process_timeout=$chainmaker_vm_go_process_timeout

  if [[ $dockervm_config_path != "" ]];then
    if [[ "${dockervm_config_path:0:1}" != "/" ]];then
        dockervm_config_path=$(pwd)/$dockervm_config_path
    fi
    if [ ! -d $mount_path/config  ];then
        mkdir $mount_path/config
    fi
    cp $dockervm_config_path $mount_path/config/vm.yml
  fi

  if [[ $protocol = "uds" ]]
  then
    docker run -itd \
    -v "$mount_path":/mount \
    -v "$log_path":/log \
    -e CHAIN_RPC_PROTOCOL="0" \
    -e MAX_SEND_MSG_SIZE="$rpc_max_send_size" \
    -e MAX_RECV_MSG_SIZE="$rpc_max_recv_size" \
    -e MAX_CONN_TIMEOUT="$rpc_timeout" \
    -e MAX_ORIGINAL_PROCESS_NUM="$max_concurrency" \
    -e DOCKERVM_CONTRACT_ENGINE_LOG_LEVEL="$vm_go_log_level" \
    -e DOCKERVM_SANDBOX_LOG_LEVEL="$vm_go_log_level" \
    -e DOCKERVM_LOG_IN_CONSOLE="$log_in_console" \
    -e SLOW_DISABLE="$slow_disable" \
    -e SLOW_STEP_TIME="$slow_step_time" \
    -e SLOW_TX_TIME="$slow_tx_time" \
    -e PROCESS_TIMEOUT="$process_timeout" \
    --name $container_name \
    --privileged $VM_GO_IMAGE_NAME \
    > /dev/null

  else
      EXPOSE_PORT=$contract_engine_port
      docker run -itd \
      --net=host \
      -v "$mount_path":/mount \
      -v "$log_path":/log \
      -e CHAIN_RPC_PROTOCOL="1" \
      -e CHAIN_HOST="$runtime_server_host" \
      -e CHAIN_RPC_PORT="$contract_engine_port" \
      -e SANDBOX_RPC_PORT="$runtime_server_port" \
      -e MAX_SEND_MSG_SIZE="$rpc_max_send_size" \
      -e MAX_RECV_MSG_SIZE="$rpc_max_recv_size" \
      -e MAX_CONN_TIMEOUT="$rpc_timeout" \
      -e MAX_ORIGINAL_PROCESS_NUM="$max_concurrency" \
      -e DOCKERVM_CONTRACT_ENGINE_LOG_LEVEL="$vm_go_log_level" \
      -e DOCKERVM_SANDBOX_LOG_LEVEL="$vm_go_log_level" \
      -e DOCKERVM_LOG_IN_CONSOLE="$log_in_console" \
      -e SLOW_DISABLE="$slow_disable" \
      -e SLOW_STEP_TIME="$slow_step_time" \
      -e SLOW_TX_TIME="$slow_tx_time" \
      -e PROCESS_TIMEOUT="$process_timeout" \
      --name $container_name \
      --privileged $VM_GO_IMAGE_NAME \
       > /dev/null
  fi

  retval="$?"
  if [ $retval -ne 0 ]; then
    echo "failed to run docker vm."
    exit 1
  fi

  echo "start docker vm service container succeed: $container_name"


  sleep 3
}

# if enable Deprecated docker vm service and use unix domain socket, it will start a docker vm container
function start_docker_vm_go() {
  container_name=DOCKERVM-{org_id}-{tagName}
  #check container exists
  exist=$(docker ps -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already RUNNING, please stop it first."
    exit 1
  fi

  exist=$(docker ps -a -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already exists(STOPPED)"
    if [[ "$FORCE_CLEAN" == "true" ]]; then
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
  docker_vm_mount_path=$chainmaker_vm_docker_go_dockervm_mount_path
  docker_vm_log_path=$chainmaker_vm_docker_go_dockervm_log_path
  docker_vm_log_level=$chainmaker_vm_docker_go_log_level
  docker_vm_log_in_console=$chainmaker_vm_docker_go_log_in_console
  if [[ "${docker_vm_mount_path:0:1}" != "/" ]];then
    docker_vm_mount_path=$(pwd)/$docker_vm_mount_path
  fi
  if [[ "${docker_vm_log_path:0:1}" != "/" ]];then
    docker_vm_log_path=$(pwd)/$docker_vm_log_path
  fi

  mkdir -p "$docker_vm_mount_path"
  mkdir -p "$docker_vm_log_path"

  # env params:
  # ENV_ENABLE_UDS=false
  # ENV_USER_NUM=1000
  # ENV_TX_TIME_LIMIT=2
  # ENV_LOG_LEVEL=INFO
  # ENV_LOG_IN_CONSOLE=false
  # ENV_MAX_CONCURRENCY=50
  # ENV_VM_SERVICE_PORT=22359
  # ENV_ENABLE_PPROF=
  # ENV_PPROF_PORT=
  docker run -itd \
    -e ENV_LOG_IN_CONSOLE="$docker_vm_log_in_console" -e ENV_LOG_LEVEL="$docker_vm_log_level" -e ENV_ENABLE_UDS=true \
    -e ENV_USER_NUM=9000 -e ENV_MAX_CONCURRENCY=100 -e ENV_TX_TIME_LIMIT=8 \
    -v "$docker_vm_mount_path":/mount \
    -v "$docker_vm_log_path":/log \
    --name $container_name \
    --privileged $DOCKER_VM_IMAGE_NAME \
    > /dev/null

  retval="$?"
  if [ $retval -ne 0 ]; then
    echo "Fail to run docker vm."
    exit 1
  fi

  echo "start deprecated docker vm service container succeed: $container_name"

  sleep 3
}

function start_vm_containers() {
    if [ "$START_FULL_MODE" == "" ]
    then
      read -r -p "start with vm containers, default: yes (y|n): " start_container
      if [ "$start_container" == "no" ] || [ "$start_container" == "n" ]; then
        START_FULL_MODE=false
      else
        START_FULL_MODE=true
      fi
    fi

    if [ "$START_FULL_MODE" == "true" ]
    then
      # check if need to start go vm service.
      if [[ "$enable_go_vm_container" == "true" ]]
      then
        start_vm_go
      fi

      # check if need to start Deprecated docker vm service.
      if [[ $enable_docker_vm_container == "true" ]]
      then
        start_docker_vm_go
      fi
    fi
}

pid=$(ps -ef | grep chainmaker | grep "\-c ../config/{org_id}/chainmaker.yml {tagName}" | grep -v grep |  awk  '{print $2}')
if [ -z "${pid}" ];then
    # check if enable go vm
    if [[ $chainmaker_vm_go_enable == "true" ]]; then
     enable_go_vm_container=true
    fi
    # check if enable Deprecated docker vm
    if [[ $chainmaker_vm_docker_go_enable_dockervm == "true" && $chainmaker_vm_docker_go_uds_open == "true" ]]; then
      enable_docker_vm_container=true
    fi
    # if enable one, start vm containers
    if [[ $enable_docker_vm_container == "true" || $enable_go_vm_container == "true" ]]
    then
      start_vm_containers
    fi

    # start chainmaker
    #nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml {tagName} > /dev/null 2>&1 &
    nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml {tagName} > panic.log 2>&1 &
    echo "chainmaker is starting, pls check log..."
else
    echo "chainmaker is already started"
fi
