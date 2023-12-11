"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""

import json

from utils.connect_linux import TheServerHelper
import config.public_import as gl


class ContractQuery(object):
    BASE_CMD = './cmc query'

    def __init__(self, sdk_config=None):
        self.new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        self.sdk_config_path = f'{gl.SDK_PATH}{self.new_sdk_config}'

    def query_block_height(self,height):
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} block-by-height {height}  --sdk-conf-path={self.sdk_config_path}'
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def query_block_hash(self,hash):
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} block-by-hash {hash}  --sdk-conf-path={self.sdk_config_path}'
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def query_block_tx_id(self,txid):
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} block-by-txid {txid}  --sdk-conf-path={self.sdk_config_path}'
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def query_by_tx(self,txid):
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} tx {txid} --sdk-conf-path={self.sdk_config_path}'
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def query_chain_config(self):
        cmd = f'cd {gl.CMC_TOOL_PATH} && ./cmc client chainconfig query --sdk-conf-path={self.sdk_config_path}'
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    # 指定在某个节点的txid查块儿高
    def query_tx_id_get_height(self):
        result = json.loads(self.query_block_tx_id())
        # print(result)
        block_height = result.get("block").get("header").get("block_height")
        return block_height

    def query_last_height(self):
        result = json.loads(self.query_block_height())
        block_height = result.get("block").get("header").get("block_height")
        print(block_height)
        return block_height

    def query_balance(self,address):
        result= self.get("balanceOf", r"[{\"address\":\"%s\"}]" % address,
             sdk_config="sdk_config.yml", abi="erc20.abi")
        balance = json.loads(result).get("contract_result").get("result")
        return balance


def query_address(cmc):
    cmd = f'cd {gl.CMC_TOOL_PATH} && ' + cmc
    print(cmd)
    result = json.loads(TheServerHelper(cmd).ssh_connectionServer())
    print(result)
    address = result.get("ethereum").get("address")
    return address

def get_user_addr(org, user):
    cmc="./cmc"
    if gl.ACCOUNT_TYPE=="cert":
        cmc=f"./cmc address cert-to-addr {gl.CRYPTO_CONFIG_PATH}/wx-org{org}.chainmaker.org/certs/user/admin1/admin1.sign.crt"
    if gl.ACCOUNT_TYPE=="pwk":
        cmc = f"./cmc address pk-to-addr {gl.CRYPTO_CONFIG_PATH}/wx-org{org}.chainmaker.org/keys/user/admin/admin.pem"
    if gl.ACCOUNT_TYPE=="pk":
        cmc = f"./cmc address pk-to-addr {gl.CRYPTO_CONFIG_PATH}/node{user}/admin/admin{user}/admin{user}.pem"
    return  query_address(cmc)
def get_user_balance(result):
    balance = json.loads(result).get("contract_result").get("result")
    return balance


if __name__ == "__main__":
    query_address("1", "client1")
    query_address("1", "admin1")
