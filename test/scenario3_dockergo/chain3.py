"""
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0

"""
import sys

sys.path.append("..")

from config.setting_chain3 import *
from testcase import *

def prepare():
    print("Gas allocation")

if __name__ == '__main__':
    UpdateSetting()
    prepare()
    unittest.main()


