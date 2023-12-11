"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import json
import sys
import unittest

sys.path.append("..")

import config.public_import as gl
from utils.cmc_tools_query import get_user_addr, get_user_balance
from utils.cmc_tools_contract import ContractDeal
from utils.cmc_command import Command
from utils.erc20 import Erc20


class Test(unittest.TestCase):
    def test_erc20_evm(self):
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
        contractNull = ContractDeal("ERC20", sync_result=True)

        result= contractNull.invoke("transfer", "{{\"k\": \"{}\",\"v\": \"{}\"}}".format("toAddr","100"),
                                 sdk_config="sdk_config.yml")
        msg=json.loads(result).get("contract_result").get("message")
        print("if invoke contract method before install ERC20 contract, response error message:",msg)
        self.assertEqual("contractName not found", msg, "success")

        print("ERC20 contract install".center(50, "="))
        cd_erc = ContractDeal("ERC20", sync_result=True)
        result_erc = cd_erc.create("EVM", "erc20.bin",abi="erc20.abi", public_identity=f'{gl.ACCOUNT_TYPE}', sdk_config='sdk_config.yml',endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}',endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',endorserOrgs=f'{gl.ADMIN_ORG_IDS}')
        erc_address = json.loads(result_erc).get("contract_result").get("result").get("address")
        print("ERC20 contract address: ", erc_address)
        erc20 = Erc20(erc_address,"erc20.abi",True,sdk_config="sdk_config.yml")
        print("withdraw contract install".center(50, "="))
        cd_withdraw = ContractDeal("withdraw", sync_result=True)
        result_withdraw = cd_withdraw.create("EVM", "withdraw.bin", public_identity=f'{gl.ACCOUNT_TYPE}',abi="withdraw.abi",
                                             endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}',endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',endorserOrgs=f'{gl.ADMIN_ORG_IDS}')
        withdraw_address = json.loads(result_withdraw).get("contract_result").get("result").get("address")
        print("withdraw contract address: ", withdraw_address)

        print("A transfer to B".center(50, "="))
        erc20.transfer(user_b_address,100)

        print("A transfer to withdraw contract".center(50, "="))
        erc20.transfer(withdraw_address,200)


        print("B invoke withdraw contract,withdraw 10".center(50, "="))
        cd_withdraw.invoke("withdraw", r'[{"address": "%s"},{"uint256": "10"}]' % erc_address,
                           sdk_config="sdk_config.yml",
                           abi="withdraw.abi", signkey=gl.USER_B_KEY,
                           signcrt="wx-org2.chainmaker.org/certs/user/admin1/admin1.sign.crt",
                           org="wx-org2.chainmaker.org")

        print("UserA balance:".center(50, "="))
        balance_a =erc20.balanceOf(user_a_address)

        expect_a = "999999999999999999999999700"

        self.assertEqual(expect_a, balance_a, "success")

        print("UserB balance:".center(50, "="))
        balance_b=erc20.balanceOf(user_b_address)
        expect_b = "110"
        self.assertEqual(expect_b, balance_b, "success")

        print("withdraw contract balance:".center(50, "="))
        balance_withdraw =erc20.balanceOf(withdraw_address)
        expect_withdraw = "190"
        self.assertEqual(expect_withdraw, balance_withdraw, "success")


if __name__ == '__main__':
    unittest.main()
