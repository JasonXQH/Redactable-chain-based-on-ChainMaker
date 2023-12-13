package mysql

import (
	"crypto/rand"
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

func generateRandomSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %v", err)
	}
	return salt, nil
}

func persistence(blockHeight uint, blockHash []byte, isModified bool) []byte {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	salt, err := generateRandomSalt()
	// 确保数据库连接成功
	if err := db.Ping(); err != nil {
		panic(err)
	}

	_, err = db.Exec("INSERT INTO block_info (block_height, block_hash, random_salt,is_modified) VALUES (?, ?,?, ?)", blockHeight, blockHash, salt, isModified)
	if err != nil {
		panic(err)
	}
	// 其他数据库操作...
	return salt
}
func getBlockInfoFromMysql(blockHeight uint) (*BlockInfo, error) {
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
