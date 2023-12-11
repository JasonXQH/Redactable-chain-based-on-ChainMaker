"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import json
import sys
import unittest
import time

sys.path.append("..")

from utils.cmc_tools_contract import ContractDeal
from utils.cmc_tools_query import ContractQuery
from utils.cmc_command import Command
import config.public_import as gl


class Test(unittest.TestCase):
    def test_invoke_native_contract(self):
        print("update ChainConfig,block interval is 20".center(50, "="))
        cmd = Command(sync_result=True)
        cmd.update_chain_config(subcmd="block updateblockinterval", params="--block-interval 20",
                                public_identity=f'{gl.ACCOUNT_TYPE}', sdk_config='sdk_config.yml',
                                endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}', endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',
                                endorserOrgs=f'{gl.ADMIN_ORG_IDS}')

        print("invoke T contract store data".center(50, "="))
        contractT = ContractDeal("T", sync_result=True)
        key = r'key%d' % time.time()
        value = r"value%d" % time.time()
        result= contractT.invoke("P", "{{\"k\": \"{}\",\"v\": \"{}\"}}".format(key,value),
                           sdk_config="sdk_config.yml")
        txId=json.loads(result).get("tx_id")
        block_height=json.loads(result).get("tx_block_height")
        # time.sleep(1)
        print("UserB query:".center(50, "="))
        result= contractT.get("G","{{\"k\":\"{}\"}}".format(key),sdk_config="sdk_config2.yml" ,stringResult=True)
        queryValue = json.loads(result).get("contract_result").get("result")
        self.assertEqual(value, queryValue, "success")

        print("UserC query:Tx".center(50, "="))
        queryCmd3 = ContractQuery(sdk_config="sdk_config3.yml")
        queryCmd3.query_by_tx(txId)
        print("UserD query:Block".center(50, "="))
        queryCmd4 = ContractQuery(sdk_config="sdk_config4.yml")
        queryCmd4.query_block_height(block_height)

        result= queryCmd4.query_chain_config()
        queryValue = json.loads(result).get("block").get("block_interval")
        self.assertEqual(20, queryValue, "success")

if __name__ == '__main__':
    unittest.main()
