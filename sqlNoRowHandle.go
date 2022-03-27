package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"log"
)

// 定义一个全局对象db
var db *sql.DB

// 定义一个初始化数据库的函数
func initDB() (err error) {
	// DSN:Data Source Name
	dsn := "root:password@tcp(127.0.0.1:3306)/dbName?charset=utf8mb4&parseTime=True"
	// 不会校验账号密码是否正确
	// 注意！！！这里不要使用:=，我们是给全局变量赋值，然后在main函数中使用全局变量db
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := initDB() // 调用输出化数据库的函数
	if err != nil {
		log.Printf("init db failed,err:%v\n", err)
		return
	}
	queryErr := queryRowDemo()

	if queryErr != nil {
		log.Println(queryErr)
	}
}

type user struct {
	id   int
	age  int
	name string
}

func queryRowDemo() error {
	sqlStr := "select id, name, age from user where id=?"
	var u user
	// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放
	err := db.QueryRow(sqlStr, 2).Scan(&u.id, &u.name, &u.age)
	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return nil
		} else {
			err1 := errors.Wrap(err, "query err")
			return err1
		}
	}

	return nil
}
