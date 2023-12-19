#!/bin/bash

# 目标目录
TARGET_DIR="/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/build/release"

# 寻找并删除所有以 'org' 结尾的文件夹
find "$TARGET_DIR" -type d -name "*.org" -exec rm -rf {} +

echo "所有以 'org' 为扩展名的文件夹已被删除。"

# MySQL 数据库凭据
MYSQL_USER="root"
MYSQL_PASSWORD="111"
MYSQL_DATABASE="chainmaker"
MYSQL_HOST="localhost"

# 清空 block_info 表
mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -h"$MYSQL_HOST" -D"$MYSQL_DATABASE" -e "TRUNCATE TABLE block_info;"

echo "block_info 表已被清空。"