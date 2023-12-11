"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""

import json

from utils.cmc_tools_contract import ContractDeal


class Erc20(object):


    def __init__(self,contract_name,abi=None, sync_result=True, sdk_config=None):
        self.contract_name=contract_name
        self.sdk_config = sdk_config
        self.contract=ContractDeal(contract_name, sync_result=sync_result)
        self.pTransfer="{{\"to\": \"{}\",\"amount\": \"{}\"}}"
        self.pBalanceOf="{{\"account\":\"{}\"}}"
        self.pApprove="{{\"spender\": \"{}\",\"amount\": \"{}\"}}"
        self.pTransferFrom="{{\"owner\": \"{}\",\"to\": \"{}\",\"amount\": \"{}\"}}"
        self.pAllowance="{{\"spender\":\"{}\",\"owner\":\"{}\"}}"
        self.stringResult=True
        self.abi=abi
        if abi:
            self.pTransfer="[{{\"address\": \"{}\"}},{{\"uint256\": \"{}\"}}]"
            self.pBalanceOf="[{{\"address\":\"{}\"}}]"
            self.pApprove="[{{\"address\": \"{}\"}},{{\"uint256\": \"{}\"}}]"
            self.pTransferFrom="[{{\"address\": \"{}\"}},{{\"address\": \"{}\"}},{{\"uint256\": \"{}\"}}]"
            self.pAllowance="[{{\"address\":\"{}\"}},{{\"address\":\"{}\"}}]"
            self.stringResult= False
    def transfer(self, to, amount):
        self.contract.invoke("transfer", self.pTransfer.format(to, amount),
                             sdk_config=self.sdk_config,abi=self.abi,stringResult=self.stringResult)

    def balanceOf(self,addr):
        result = self.contract.get("balanceOf", self.pBalanceOf.format( addr),
                                   sdk_config=self.sdk_config, abi=self.abi,stringResult=self.stringResult)
        balance = json.loads(result).get("contract_result").get("result")
        if self.abi:
            return balance[1:-1]
        return balance

    def approve(self, spender, amount):
        self.contract.invoke("approve", self.pApprove.format (spender, amount),
                             sdk_config=self.sdk_config,abi=self.abi,stringResult=self.stringResult)

    def transferFrom(self,_from, to, amount):
        self.contract.invoke("transferFrom", self.pTransferFrom.format (_from,to, amount),
                             sdk_config=self.sdk_config,abi=self.abi,stringResult=self.stringResult)

    def allowance(self,owner,spender):
        result = self.contract.get("allowance", self.pAllowance.format (owner,spender),
                                   sdk_config=self.sdk_config, abi=self.abi,stringResult=self.stringResult)
        allowance_amt = json.loads(result).get("contract_result").get("result")
        if self.abi:
            return allowance_amt[1:-1]
        return allowance_amt