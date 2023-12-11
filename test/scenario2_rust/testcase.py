"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import base64
import json
import sys
import unittest

sys.path.append("..")

import config.public_import as gl
from utils.cmc_tools_contract import ContractDeal
from utils.cmc_tools_query import get_user_addr
from utils.cmc_command import Command


class Test(unittest.TestCase):
    def test_balance_a_compare_cert(self):
        print("query UserA address: org1 admin".center(50, "="))
        user_a_address = get_user_addr("1", "1")
        print("query UserB address: org2 admin".center(50, "="))
        user_b_address = get_user_addr("2", "2")
        print("query UserC address: org3 admin".center(50, "="))
        user_c_address = get_user_addr("3", "3")
        print("query UserD address: org4 admin".center(50, "="))
        user_d_address = get_user_addr("4", "4")
        print("User ABCD address:", user_a_address, user_b_address, user_c_address, user_d_address)

        if gl.ENABLE_GAS == True:
            cmd = Command(sync_result=True)
            cmd.recharge_gas(user_a_address)
            cmd.recharge_gas(user_b_address)
            cmd.recharge_gas(user_c_address)
            cmd.recharge_gas(user_d_address)

        print("\n","rust asset contract install".center(50, "="))
        cd_asset = ContractDeal("asset", sync_result=True)
        result_erc = cd_asset.create("WASMER", "rust-asset-2.0.0.wasm",
                                     params=r"{\"issue_limit\":\"10000000\",\"total_supply\":\"1000000000\"}",
                                     public_identity=f'{gl.ACCOUNT_TYPE}', sdk_config='sdk_config.yml',
                                     endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}',endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',
                                     endorserOrgs=f'{gl.ADMIN_ORG_IDS}')
        asset_address = json.loads(result_erc).get("contract_result").get("result").get("address")
        print("rust asset contract address:",asset_address,"\n")


        print("register B account".center(50, "="))
        user_b_address_result = cd_asset.invoke("register", "",sdk_config="sdk_config2.yml")
        user_b_address = str(base64.b64decode(json.loads(user_b_address_result).get("contract_result").get("result")),encoding='utf-8')


        print("register C account".center(50, "="))
        user_c_address_result = cd_asset.invoke("register", "",sdk_config="sdk_config3.yml")
        user_c_address = str(base64.b64decode(json.loads(user_c_address_result).get("contract_result").get("result")),encoding='utf-8')



        print("query UserA address: org1 admin".center(50, "="))
        user_a_address_result = cd_asset.get("query_address", "", sdk_config="sdk_config.yml")
        user_a_address = str(base64.b64decode(json.loads(user_a_address_result).get("contract_result").get("result")),encoding='utf-8')


        print("query UserB address: org2 admin".center(50, "="))
        user_b_address_result2 = cd_asset.get("query_address", "", sdk_config="sdk_config2.yml")
        user_b_address2 = str(base64.b64decode(json.loads(user_b_address_result2).get("contract_result").get("result")),encoding='utf-8')
        self.assertEqual(user_b_address2, user_b_address, "success")



        print("query UserC address: org3 admin".center(50, "="))
        user_c_address_result2 = cd_asset.get("query_address", "", sdk_config="sdk_config3.yml")
        user_c_address2 = str(base64.b64decode(json.loads(user_c_address_result2).get("contract_result").get("result")),encoding='utf-8')
        self.assertEqual(user_c_address2, user_c_address, "success")


        print("\n","User A address:",user_a_address,"\n","User B address:",user_b_address2,"\n","User C address:",user_c_address2,"\n")



        print("issue A account token 100".center(50, "="))
        cd_asset.invoke("issue_amount", "{{\"to\": \"{}\",\"amount\": \"{}\"}}".format(user_a_address,100),
                        sdk_config="sdk_config.yml")


        print("issue B account token 100".center(50, "="))
        cd_asset.invoke("issue_amount", "{{\"to\":\"{}\",\"amount\":\"{}\"}}".format(user_b_address2,100),
                        sdk_config="sdk_config.yml")


        print("A transfer to B 10".center(50, "="))
        cd_asset.invoke("transfer", "{{\"to\":\"{}\",\"amount\":\"{}\"}}".format(user_b_address2,10),
                        sdk_config="sdk_config.yml")


        print("B allowance A 50 token".center(50, "="))
        cd_asset.invoke("approve", "{{\"spender\":\"{}\",\"amount\":\"{}\"}}".format(user_a_address,50),sdk_config="sdk_config2.yml")




        print("A transfer to C with balance of B allowance A".center(50, "="))
        cd_asset.invoke("transfer_from", "{{\"from\":\"{}\",\"to\":\"{}\",\"amount\":\"{}\"}}".format(user_b_address2,user_c_address2,10),
                        sdk_config="sdk_config.yml")



        print("query balance of B allowance A,should be 40".center(50, "="))
        balance_b_allowance_a_result = cd_asset.get("allowance",
                                                    "{{\"owner\":\"{}\",\"spender\":\"{}\"}}".format(user_b_address2,user_a_address),
                                                    sdk_config="sdk_config.yml")

        balance_b_allowance_a = base64.b64decode(json.loads(balance_b_allowance_a_result).get("contract_result").get("result"))
        print("query result:balance of B allowance A:",balance_b_allowance_a,"\n")

        print("query A account balance,should be 90".center(50, "="))
        balance_a_result = cd_asset.get("balance_of",
                                        "{{\"owner\":\"{}\"}}".format(user_a_address),
                                        sdk_config="sdk_config.yml", signkey="", signcrt="", org="")
        balance_user_a = base64.b64decode(json.loads(balance_a_result).get("contract_result").get("result"))
        print("query result:balance of A:",balance_user_a,"\n")


        print("query B account balance,should be 100".center(50, "="))
        balance_b_result = cd_asset.get("balance_of","{{\"owner\":\"{}\"}}".format(user_b_address2),
                                        sdk_config="sdk_config2.yml")



        balance_user_b = base64.b64decode(json.loads(balance_b_result).get("contract_result").get("result"))
        print("query result:balance of B",balance_user_b,"\n")



if __name__ == '__main__':
    unittest.main()

