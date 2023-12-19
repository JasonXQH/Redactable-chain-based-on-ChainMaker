#!/bin/bash

# 运行第一条命令
./cmc client contract user create \
--contract-name=fact \
--runtime-type=WASMER \
--byte-code-path=./testdata/claim-wasm-demo/rust-fact-2.0.0.wasm \
--version=1.0 \
--sdk-conf-path=./testdata/sdk_config.yml \
--admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.key,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.key \
--admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org2.chainmaker.org/user/admin1/admin1.sign.crt,./testdata/crypto-config/wx-org3.chainmaker.org/user/admin1/admin1.sign.crt \
--sync-result=true \
--params="{}"

# 检查命令是否成功执行
if [ $? -ne 0 ]; then
    echo "第一条命令执行失败"
    exit 1
fi

# 运行第二条命令
./cmc client contract user invoke \
--contract-name=fact \
--method=save \
--sdk-conf-path=./testdata/sdk_config.yml \
--params="{\"file_name\":\"name007\",\"file_hash\":\"ab3456df5799b87c77e7f88\",\"time\":\"6543234\"}" \
--sync-result=true

# 检查命令是否成功执行
if [ $? -ne 0 ]; then
    echo "第二条命令执行失败"
    exit 1
fi

# 运行第三条命令
./cmc query block-by-height 1 \
--chain-id=chain1 \
--sdk-conf-path=./testdata/sdk_config.yml

# 检查命令是否成功执行
if [ $? -ne 0 ]; then
    echo "第三条命令执行失败"
    exit 1
fi
