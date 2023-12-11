"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import os

import config.public_import as gl

def UpdateSetting():
    gl.TESTPROJECTPATH = os.path.abspath(os.getcwd() + r"/../chain3/bin")
    gl.CMC_TOOL_PATH = gl.TESTPROJECTPATH
    gl.SDK_CONFIG_PATH = r'../config/sdk_config.yml'
    gl.CRYPTO_CONFIG_PATH = r'../config'
    gl.ADMIN_KEY_FILE_PATHS = ','.join([f'{gl.CRYPTO_CONFIG_PATH}/node{i}/admin/admin{i}/admin{i}.key'
                                        for i in range(1, 4)])

    gl.WASM_APTH = r'../../testdata/'
    gl.SDK_PATH = r'../config/'
    gl.ACCOUNT_TYPE = "pk"
    gl.USER_B_KEY = r"node2/admin/admin2/admin2.key"
    gl.ENABLE_GAS = True
