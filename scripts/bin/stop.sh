#!/bin/bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# read params
for i in "$@";do
    case "$i" in
        alone)
            STOP_FULL_MODE=false
            ;;
        full)
            STOP_FULL_MODE=true
            ;;
        *)
            echo $"Usage: $0 {alone|full}"
            echo $"  full: both stop chainmaker node and vm containers"
            echo $"  alone: stop chainmaker node, WITHOUT vm containers"
            exit
    esac
done


# stop chainmaker node
pid=`ps -ef | grep chainmaker | grep "\-c ../config/{org_id}/chainmaker.yml {tagName}" | grep -v grep |  awk  '{print $2}'`
if [ ! -z ${pid} ];then
    kill $pid
    echo "chainmaker is stopping..."
fi


# check if need to stop vm containers
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

function stop_container() {
  container_name=$1
  container_exists=$(docker ps -f name="$container_name" --format '{{.Names}}')

  if [[ $container_exists ]]; then
    echo -n "stop docker vm container: "
    docker stop "$container_name"
  fi
}

config_file="../config/{org_id}/chainmaker.yml"
# config_file="../../config/wx-org1-solo/chainmaker.yml"
eval $(parse_yaml "$config_file" "chainmaker_")

go_vm_container_name=VM-GO-{org_id}-{tagName}
# Deprecated docker vm container
docker_vm_container_name=DOCKERVM-{org_id}-{tagName}

if [[ $chainmaker_vm_go_enable == "true" ]]; then
 enable_go_vm_container=true
fi
# check if enable Deprecated docker vm
if [[ $chainmaker_vm_docker_go_enable_dockervm == "true" && $chainmaker_vm_docker_go_uds_open == "true" ]]; then
  enable_docker_vm_container=true
fi

# if enable one, stop vm containers
if [[ $enable_docker_vm_container == "true" || $enable_go_vm_container == "true" ]]
then
  if [ "$STOP_FULL_MODE" == "" ]
  then
    read -r -p "stop vm containers, default: yes (y|n): " stop_container
    if [ "$stop_container" == "no" ] || [ "$stop_container" == "n" ]; then
      STOP_FULL_MODE=false
    else
      STOP_FULL_MODE=true
    fi
  fi

  if [ "$STOP_FULL_MODE" == "true" ]
  then
    # if enable go vm service, stop the running container
    if [[ $enable_go_vm_container == "true" ]]
    then
      stop_container $go_vm_container_name
    fi

    # if enable Deprecated docker vm service and use unix domain socket, stop the running container
    if [[ $enable_docker_vm_container == "true" ]]
    then
      stop_container $docker_vm_container_name
    fi
  fi
fi


if [ ! -z "${pid}" ];then
    lsof -p "$pid" +r 1 &>/dev/null
fi
echo "chainmaker is stopped"
