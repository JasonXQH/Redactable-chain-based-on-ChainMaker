"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import base64
import json
import string
import sys
import unittest
import re

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

        print("\n","1.rust asset contract install".center(50, "="))
        cd_asset = ContractDeal("rustsql", sync_result=True)
        result_erc = cd_asset.create("WASMER", "rust-sql-2.0.0.wasm",
                                     public_identity=f'{gl.ACCOUNT_TYPE}', sdk_config='sdk_config.yml',
                                     endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}',endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',
                                     endorserOrgs=f'{gl.ADMIN_ORG_IDS}')
        asset_address = json.loads(result_erc).get("contract_result").get("result").get("address")
        print("rust asset contract address:",asset_address,"\n")


        print("2.invoke contract-sql insert".center(50, "="))
        for i in range(1,10):
            cd_asset.invoke("sql_insert",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\",\"age\":\"{}\",\"id_card_no\":\"{}\"}}".format(str(i),"chainmaker",str(i+10),"510623199202023323"))



        print("3.query record of age is 11".center(50, "="))
        query_result=cd_asset.invoke("sql_query_by_id",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(1)))
        query=base64.b64decode(json.loads(query_result).get("contract_result").get("result"))
        print("id:",str(base64.b64decode(json.loads(query).get("id")),encoding='utf-8'))
        print("name:",str(base64.b64decode(json.loads(query).get("name")),encoding='utf-8'))
        print("age:",str(base64.b64decode(json.loads(query).get("age")),encoding='utf-8'))
        print("id_card_no:",str(base64.b64decode(json.loads(query).get("id_card_no")),encoding='utf-8'),"\n")


        print("4.invoke sql sentence: update name=chainmaker_update where id=1".center(50, "="))
        cd_asset.invoke("sql_update",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(1),"chainmaker_update"))



        print("5.query if the name of id=1 update to chainmaker_update".center(50, "="))
        name_update_result=cd_asset.get("sql_query_by_id",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(1)))
        name_update=base64.b64decode(json.loads(name_update_result).get("contract_result").get("result"))
        print("id:",str(base64.b64decode(json.loads(name_update).get("id")),encoding='utf-8'))
        print("name:",str(base64.b64decode(json.loads(name_update).get("name")),encoding='utf-8'))
        print("age:",str(base64.b64decode(json.loads(name_update).get("age")),encoding='utf-8'))
        print("id_card_no:",str(base64.b64decode(json.loads(name_update).get("id_card_no")),encoding='utf-8'),"\n")


        print("6.range query: rang age 1~10".center(50, "="))
        range_age_result=cd_asset.invoke("sql_query_range_of_age",sdk_config="sdk_config.yml",params="{{\"min_age\":\"{}\",\"max_age\":\"{}\"}}".format(str(13),str(17)))
        range_age=str(base64.b64decode(json.loads(range_age_result).get("contract_result").get("result")),encoding='utf-8')
        parts_range_age=re.split('{|}',range_age)
        for part_range_age in parts_range_age:
            if part_range_age=='':
                continue
            part_range_age='{'+part_range_age+'}'
            print("id:",str(base64.b64decode(json.loads(part_range_age).get("id")),encoding='utf-8'))
            print("name:",str(base64.b64decode(json.loads(part_range_age).get("name")),encoding='utf-8'))
            print("age:",str(base64.b64decode(json.loads(part_range_age).get("age")),encoding='utf-8'))
            print("id_card_no:",str(base64.b64decode(json.loads(part_range_age).get("id_card_no")),encoding='utf-8'),"\n")


        print("7.invoke contract-sql delete by id age=11".center(50, "="))
        cd_asset.invoke("sql_delete",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(1)))


        print("8.query id age=11 once more,should not found".center(50, "="))
        query_id_result=cd_asset.get("sql_query_by_id",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(1)))
        query_id=str(base64.b64decode(json.loads(query_id_result).get("contract_result").get("result")),encoding='utf-8')
        print("query id age=11 once more,the result is: ",query_id,"\n")
        self.assertEqual(query_id, '{}', "success")


        print("9.cross invoke contract".center(50, "="))
        sql_cross_call_result=cd_asset.get("sql_cross_call",sdk_config="sdk_config.yml",params="{{\"contract_name\":\"{}\",\"min_age\":\"{}\",\"max_age\":\"{}\"}}".format("rustsql",str(16),str(19)))
        sql_cross_call=str(base64.b64decode(json.loads(sql_cross_call_result).get("contract_result").get("result")),encoding='utf-8')
        parts_range_age=re.split('{|}',sql_cross_call)
        for part_range_age in parts_range_age:
            if part_range_age=='':
                continue
            part_range_age='{'+part_range_age+'}'
            print("id:",str(base64.b64decode(json.loads(part_range_age).get("id")),encoding='utf-8'))
            print("name:",str(base64.b64decode(json.loads(part_range_age).get("name")),encoding='utf-8'))
            print("age:",str(base64.b64decode(json.loads(part_range_age).get("age")),encoding='utf-8'))
            print("id_card_no:",str(base64.b64decode(json.loads(part_range_age).get("id_card_no")),encoding='utf-8'),"\n")


        print("10.transaction revert".center(50, "="))
        cd_asset.invoke("sql_insert",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\",\"age\":\"{}\",\"id_card_no\":\"{}\"}}".format(str(20),"chainmaker",str(2000),"510623199202023323"))

        print("10.1 commit a failed transaction".center(50, "="))
        cd_asset.invoke("sql_update_rollback_save_point",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(20),"chainmaker_save_point"))

        print("10.2 query the failed transaction affect to the last transaction".center(50, "="))
        query_id_result=cd_asset.get("sql_query_by_id",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(20)))
        query_id=str(base64.b64decode(json.loads(query_id_result).get("contract_result").get("result")),encoding='utf-8')

        print("id:",str(base64.b64decode(json.loads(query_id).get("id")),encoding='utf-8'))
        print("name:",str(base64.b64decode(json.loads(query_id).get("name")),encoding='utf-8'))
        print("age:",str(base64.b64decode(json.loads(query_id).get("age")),encoding='utf-8'))
        print("id_card_no:",str(base64.b64decode(json.loads(query_id).get("id_card_no")),encoding='utf-8'),"\n")

        name=str(base64.b64decode(json.loads(query_id).get("name")),encoding='utf-8')
        self.assertEqual(name, 'chainmaker', "success")


        print("11.upgrade contract".center(50, "="))
        result_erc = cd_asset.upgrade("WASMER", "rust-sql-2.0.0.wasm",
                                      public_identity=f'{gl.ACCOUNT_TYPE}', sdk_config='sdk_config.yml',version="2.0.1",
                                      endorserKeys=f'{gl.ADMIN_KEY_FILE_PATHS}',endorserCerts=f'{gl.ADMIN_CRT_FILE_PATHS}',
                                      endorserOrgs=f'{gl.ADMIN_ORG_IDS}')
        asset_address = json.loads(result_erc).get("contract_result").get("result").get("address")
        print("rust asset contract address:",asset_address,"\n")


        print("12.invoke insert after upgrade contract".center(50, "="))
        cd_asset.invoke("sql_insert",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\",\"age\":\"{}\",\"id_card_no\":\"{}\"}}".format(str(21),"chainmaker",str(100000),"510623199202023323"))

        query_id_result=cd_asset.get("sql_query_by_id",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\"}}".format(str(21)))
        query_id=str(base64.b64decode(json.loads(query_id_result).get("contract_result").get("result")),encoding='utf-8')

        print("id:",str(base64.b64decode(json.loads(query_id).get("id")),encoding='utf-8'))
        print("name:",str(base64.b64decode(json.loads(query_id).get("name")),encoding='utf-8'))
        print("age:",str(base64.b64decode(json.loads(query_id).get("age")),encoding='utf-8'))
        print("id_card_no:",str(base64.b64decode(json.loads(query_id).get("id_card_no")),encoding='utf-8'),"\n")

        age=str(base64.b64decode(json.loads(query_id).get("age")),encoding="utf-8")
        self.assertEqual(age,str(100000),"success")


        print("13.parallel testing".center(50, "="))
        for i in range(500,600):
            cd_asset2 = ContractDeal("rustsql", sync_result=False)
            cd_asset2.invoke("sql_insert",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\",\"age\":\"{}\",\"id_card_no\":\"{}\"}}".format(str(i),"chainmaker",str(i+10),"510623199202023323"))



        print("14.abnormal function test".center(50, "="))
        print("14.1 create table,index,view and so on DDL sentence can only use in init_contract and upgrade contract".center(50, "="))
        ddl_result=cd_asset.invoke("sql_execute_ddl",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        ddl_message=json.loads(ddl_result).get("contract_result").get("message");
        b="match expection" in ddl_message
        if b:
            print("result contains match expectio pass!!!\n")


        print("14.2 in SQL, forbid cross database operate,without assign database name.such as select * from db.table is forbidden； use db; is forbidden.".center(50, "="))
        forbidden_result=cd_asset.invoke("sql_dbname_table_name",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        forbidden_message=json.loads(forbidden_result).get("contract_result").get("message");
        b="match expection" in forbidden_message
        if b:
            print("result contains match expection pass!!!\n")


        print("14.3 in SQL,forbid using transaction operation sentence,such as commit ,rollback,transaction is controlled by ChainMaker framework.".center(50, "="))
        tx_result=cd_asset.invoke("sql_execute_commit",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        tx_message=json.loads(tx_result).get("contract_result").get("message");
        b="match expection" in tx_message
        if b:
            print("result contains match expectio pass!!!\n")


        print("14.4 in SQL,forbid using random number,getting system time and so on uncertain function,these function result is different from different node result can not consensus.".center(50, "="))
        random_key_result=cd_asset.invoke("sql_random_key",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        random_key=json.loads(random_key_result).get("contract_result").get("message");
        b="forbidden sql keyword" in random_key
        if b:
            print("result contains forbidden sql keyword, pass!!!\n")

        random_str_result=cd_asset.invoke("sql_random_str",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(502),"chainmaker"))
        random_str=str(base64.b64decode(json.loads(random_str_result).get("contract_result").get("result")),encoding='utf-8')
        self.assertEqual(random_str,"ok","success")
        print("result:",random_str, "pass !!!\n")

        random_query_str_result=cd_asset.get("sql_random_query_str",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        random_query_str=str(base64.b64decode(json.loads(random_query_str_result).get("contract_result").get("result")),encoding='utf-8')
        self.assertEqual(random_query_str,"ok","success")
        print("result:",random_query_str, "pass !!!\n")

        print("14.5 in SQL,forbid multi SQL combination one SQL string to input ".center(50, "="))
        multi_sql_result=cd_asset.invoke("sql_multi_sql",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        multi_sql=json.loads(multi_sql_result).get("contract_result").get("message");
        b="match expection" in multi_sql
        if b:
            print("result contains match expection pass!!!\n")


        print("14.6 forbid create,fix,delete the table named state_infos”,this is system provided kv data table,using to store PutState function data.".center(50, "="))
        update_state_info_result=cd_asset.invoke("sql_update_state_info",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        update_state_info=json.loads(update_state_info_result).get("contract_result").get("message");
        b="you can't change table state_infos" in update_state_info
        if b:
            print("result contains you can't change table state_infos","pass!!!\n")

        query_state_info_result=cd_asset.get("sql_query_state_info",sdk_config="sdk_config.yml",params="{{\"id\":\"{}\",\"name\":\"{}\"}}".format(str(501),"chainmaker"))
        query_state_info=json.loads(query_state_info_result).get("contract_result").get("message");
        b="match expection" in query_state_info
        if b:
            print("result contains match expection pass!!!\n")



if __name__ == '__main__':
    unittest.main()