"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""

import config.public_import as gl
from utils.connect_linux import TheServerHelper


class ContractDeal(object):
    BASE_CMD = './cmc client contract user'

    def __init__(self, contract_name, sync_result=True):
        """
        :param contract_name: 合约名称
        :param sync_result: 同步参数,默认true,传入false就异步
        """
        self.contract_name = contract_name
        self.sync_result = sync_result

    def create(self, runtime, wasm, abi=None, params=None, public_identity=None, sdk_config=None,endorserKeys =None,endorserCerts=None,endorserOrgs=None):
        """
        支持创建普通sql合约以及kv合约
        :param runtime: GASM,WASMER,DOCKER_GO,EVM
        :param wasm: 使用的是哪个wasm文件
        :param params: 创建合约的时候的参数
        :param public_identity: 是否选择公钥方式,共有三种模式,pk,pwk,cert模式
        :param sdk_config: 默认是节点1,传入的这个sdk的配置文件表示在哪个节点上跑
        :return:
        """
        wasm_path = gl.WASM_APTH + wasm
        # params_new = params if params else "{}"
        new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        # new_abi = abi if abi else wasm.split(".")[0] + ".abi"
        sdk_config_path = f'{gl.SDK_PATH}{new_sdk_config}'
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} create --contract-name={self.contract_name} --runtime-type={runtime} --byte-code-path={wasm_path} --version=1.0 --sdk-conf-path={sdk_config_path}'
        if public_identity == "pwk":
            print("contract create-pwk mode".center(50, "="))
            cmd = cmd + f' --admin-org-ids={endorserOrgs} --admin-key-file-paths={endorserKeys}'
        elif public_identity == "pk":
            print("contract create-pk mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys}'
        else:
            print("contract create-cert mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys} --admin-crt-file-paths={endorserCerts}'
        if gl.ENABLE_GAS:
            cmd = cmd + " --gas-limit=99999999"
        if params :
            cmd = cmd + f' --params="{params}"'
        if abi :
            cmd = cmd + f" --abi-file-path={gl.WASM_APTH}{abi}"
        if self.sync_result:
            cmd = cmd + " --sync-result=true"
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def invoke(self, method, params, sdk_config=None, txid=None, abi=None, signkey=None, signcrt=None, org=None, stringResult=False):
        """
        调用合约
        :param method: 调用合约方法
        :param params: 调用合约的参数
        :param org: 指定在某个节点上执行
        :return: 返回合约调用结果
        """
        new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        sdk_config_path = f'{gl.SDK_PATH}{new_sdk_config}'
        cmd =f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} invoke --contract-name={self.contract_name} --method={method} --sdk-conf-path={sdk_config_path}'
        if txid :
            cmd = cmd + f" --tx-id={txid}"
        if signkey:
            cmd = cmd + f" --user-signkey-file-path={gl.CRYPTO_CONFIG_PATH}/{signkey}"
        if signcrt:
            cmd = cmd + f" --user-signcrt-file-path={gl.CRYPTO_CONFIG_PATH}/{signcrt}"
        if abi:
            cmd = cmd + f' --abi-file-path={gl.WASM_APTH}{abi}'
        if params:
            cmd = cmd + f' --params=\'{params}\''
        if self.sync_result:
            cmd = cmd + f' --sync-result=true'
        if org:
            cmd = cmd + f" --org-id={org}"

        if gl.ENABLE_GAS:
            cmd = cmd + " --gas-limit=99999999"
        if stringResult:
            cmd = cmd + " --result-to-string=true"
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def get(self, method, params, sdk_config=None, abi=None, stringResult=False, signkey=None, signcrt=None, org=None):
        """
        查询合约
        :param method: 查询合约的方法
        :param params: 查询合约的参数
        :return:
        """
        new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        sdk_config_path = f'{gl.SDK_PATH}{new_sdk_config}'
        cmd =f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} get --contract-name={self.contract_name} --method={method} --sdk-conf-path={sdk_config_path}'
        if abi:
            cmd = cmd + f' --abi-file-path={gl.WASM_APTH}{abi}'
        if params:
            cmd = cmd + f' --params=\'{params}\''
        if stringResult:
            cmd = cmd + " --result-to-string=true"
        if signkey:
            cmd = cmd + f" --user-signkey-file-path={gl.CRYPTO_CONFIG_PATH}/{signkey}"
        if signcrt:
            cmd = cmd + f" --user-signcrt-file-path={gl.CRYPTO_CONFIG_PATH}/{signcrt}"
        if org:
            cmd = cmd + f" --org-id={org}"
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result



    def upgrade(self, runtime, wasm, abi=None, version=None,params=None, public_identity=None, sdk_config=None,endorserKeys =None,endorserCerts=None,endorserOrgs=None):
        """
        支持升级普通sql合约以及kv合约
        :param runtime: GASM,WASMER,DOCKER_GO,EVM
        :param wasm: 使用的是哪个wasm文件
        :param params: 升级合约的时候的参数
        :param public_identity: 是否选择公钥方式,共有三种模式,pk,pwk,cert模式
        :param sdk_config: 默认是节点1,传入的这个sdk的配置文件表示在哪个节点上跑
        :return:
        """
        wasm_path = gl.WASM_APTH + wasm
        # params_new = params if params else "{}"
        new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        # new_abi = abi if abi else wasm.split(".")[0] + ".abi"
        sdk_config_path = f'{gl.SDK_PATH}{new_sdk_config}'
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} upgrade --contract-name={self.contract_name} --runtime-type={runtime} --byte-code-path={wasm_path} --version={version} --sdk-conf-path={sdk_config_path}'
        if public_identity == "pwk":
            print("contract create-pwk mode".center(50, "="))
            cmd = cmd + f' --admin-org-ids={endorserOrgs} --admin-key-file-paths={endorserKeys}'
        elif public_identity == "pk":
            print("contract create-pk mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys}'
        else:
            print("contract create-cert mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys} --admin-crt-file-paths={endorserCerts}'
        if gl.ENABLE_GAS:
            cmd = cmd + " --gas-limit=99999999"
        if params :
            cmd = cmd + f' --params="{params}"'
        if abi :
            cmd = cmd + f" --abi-file-path={gl.WASM_APTH}{abi}"
        if self.sync_result:
            cmd = cmd + " --sync-result=true"
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result





if __name__ == "__main__":
    # cd = ContractDeal("feifei_test_001", sync_result=True)
    # cd.create("WASMER", "rust-fact-2.0.0.wasm", public_identity='pk', sdk_config='sdk_config_pk.yml')
    # cd.create("WASMER", "rust-fact-2.0.0.wasm")
    # cd.create("WASMER", "chainmaker_contract_no_upgrade.wasm")
    # cd.create("WASMER", "rust-asset-2.1.0.wasm",
    #           params=r'{\"issue_limit\":\"100\",\"total_supply\":\"1000\",\"manager_pk\":\"\"}')
    # print(cd.create("WASMER",r'{\"status\":\"success\",\"tblinfo\":\"CREAE TABLE Persons(Id_P int,LastName varchar(255),FirstName varchar(255),Address varchar(255),City varchar(255))\"}'))
    # cd.invoke("register", '', org=2)

    # cd.invoke("save",
    #                 r'{\"file_name\":\"name007\",\"file_hash\":\"ab3456df5799b87c77e7f88\",\"time\":\"6543234\"}')
    # cd.get("query_address", '')
    # cd.get("balance_of", r'{\"owner\":\"2d0e03297ff63ce802d2b8a71ee8efe17001f6c9da1816cf15540c982849520b\"}')
    # print(cd.upgrade("WASMER", "2.0", "chainmaker_contract.wasm"))
    # print(cd.freeze())
    # print(cd.unfreeze())
    # print(cd.revoke())
    cd = ContractDeal("ERC20", sync_result=True)
    a = cd.get("balanceOf", r"[{\"address\":\"d8551a6f75e0b76cb22d1e0a2770355396f66963\"}]", abi="erc20.abi")
    print(type(a))
