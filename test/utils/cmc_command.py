"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""

import json

from utils.connect_linux import TheServerHelper
import config.public_import as gl


class Command(object):
    BASE_CMD = './cmc '

    def __init__(self, sync_result=True, sdk_config=None):
        self.sync_result = sync_result
        self.new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        self.sdk_config_path = f'{gl.SDK_PATH}{self.new_sdk_config}'

    def recharge_gas(self, address, amount=1000000000):
        cmd =   f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} gas recharge --address={address} --amount={amount} --sdk-conf-path={self.sdk_config_path}'
        if self.sync_result:
            cmd = cmd + " --sync-result=true"
        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

    def update_chain_config(self, subcmd, params=None, public_identity=None, sdk_config=None,endorserKeys =None,endorserCerts=None,endorserOrgs=None):
        new_sdk_config = sdk_config if sdk_config else "sdk_config.yml"
        sdk_config_path = f'{gl.SDK_PATH}{new_sdk_config}'
        cmd = f'cd {gl.CMC_TOOL_PATH} && {self.BASE_CMD} client chainconfig {subcmd} {params} --sdk-conf-path={sdk_config_path}'
        if public_identity == "pwk":
            print("contract create-pwk mode".center(50, "="))
            cmd = cmd + f' --admin-org-ids={endorserOrgs} --admin-key-file-paths={endorserKeys}'
        elif public_identity == "pk":
            print("contract create-pk mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys}'
        else:
            print("contract create-cert mode".center(50, "="))
            cmd = cmd + f' --admin-key-file-paths={endorserKeys} --admin-crt-file-paths={endorserCerts}'

        print(cmd)
        result = TheServerHelper(cmd).ssh_connectionServer()
        print(result)
        return result

