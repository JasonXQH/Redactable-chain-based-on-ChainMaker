#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
docker rm -f mysql
docker run --name mysql -e MYSQL_ROOT_PASSWORD=123 -p 3306:3306 -d mysql
