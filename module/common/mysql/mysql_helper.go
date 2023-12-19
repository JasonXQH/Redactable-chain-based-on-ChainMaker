package mysql

import (
	"chainmaker.org/chainmaker-go/module/common"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type BlockInfo struct {
	BlockHeight uint
	BlockHash   []byte
	RandomSalt  []byte
	IsModified  bool
}

const (
	mysqlUser     = "root"       // MySQL用户名
	mysqlPassword = "111"        // MySQL密码
	mysqlHost     = "localhost"  // MySQL主机
	mysqlPort     = "3306"       // MySQL端口
	mysqlDatabase = "chainmaker" // MySQL数据库名
)

func Persistence(block_height uint64, merkleTreeRoot common.Hash, salt []byte) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// 确保数据库连接成功
	if err := db.Ping(); err != nil {
		panic(err)
	}
	_, err = db.Exec("INSERT INTO block_info (block_height, merkletree_root, random_salt,is_modified) VALUES (?, ?,?, ?)", block_height, merkleTreeRoot, salt, false)
	if err != nil {
		panic(err)
	}
	// 其他数据库操作...
}

func UpdateSalt(info *BlockInfo) error {
	// 连接到数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// 更新数据库中的salt
	query := "UPDATE block_info SET random_salt = ? WHERE block_height = ?"
	result, err := db.Exec(query, info.RandomSalt, info.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to update salt in database: %v", err)
	}

	// 检查是否真的更新了数据库中的条目
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated, block height %d may not exist", info.BlockHeight)
	}
	return nil
}

func GetSalt(blockheight uint64) ([]byte, error) {
	// 连接到数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// 更新数据库中的salt
	query := "SELECT random_salt FROM block_info WHERE block_height = ?"
	var salt []byte
	err = db.QueryRow(query, blockheight).Scan(&salt)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有找到对应的条目
			return nil, nil
		}
		// 数据库查询出错
		return nil, fmt.Errorf("failed to query randomsalt from database: %v", err)
	}
	return salt, nil
}

func GetBlockInfoFromMysql(blockHeight uint) (*BlockInfo, error) {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 查询语句
	query := "SELECT block_height, block_hash, is_modified FROM block_info WHERE block_height = ?"

	// 执行查询
	var blockInfo BlockInfo
	err = db.QueryRow(query, blockHeight).Scan(&blockInfo.BlockHeight, &blockInfo.BlockHash, &blockInfo.IsModified)
	// 检查是否未找到条目
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no entry found for block height %d", blockHeight)
	} else if err != nil {
		return nil, err // 处理其他潜在的错误
	}
	fmt.Println("BlockHeight: ", blockInfo.BlockHeight, " BlockHash: ", blockInfo.BlockHash, " IsModified: ", blockInfo.IsModified)
	return &blockInfo, nil
}
