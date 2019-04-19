package orm

import "database/sql"
import (
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type DbDriverMysql struct {
	db *sql.DB
}

var dbPoolMap = map[string]*sql.DB{}
var dbPoolMapLock *sync.RWMutex = &sync.RWMutex{}

// 单例取出db 并返回自己
// err 是打开数据库连接的错误
func Singleton(connect *Connect) (*DbDriverMysql, error) {
	configString := connect.String()

	dbPoolMapLock.RLock()
	db, isOk := dbPoolMap[configString]

	if !isOk {
		dbPoolMapLock.RUnlock()
		dbPoolMapLock.Lock()
		_db, err := sql.Open(connect.Driver, connect.SqlString())
		if err != nil {
			return nil, err
		}
		_db.SetMaxOpenConns(2000)

		db = _db
		dbPoolMap[configString] = db
		dbPoolMapLock.Unlock()
	} else {
		err := db.Ping()
		if err != nil {
			// 如果ping不通,就删除这个连接,报错
			delete(dbPoolMap, configString)
			dbPoolMapLock.RUnlock()
			return nil, err
		}
		dbPoolMapLock.RUnlock()
	}

	dbDriverMysql := DbDriverMysql{}
	dbDriverMysql.db = db

	return &dbDriverMysql, nil
}

// 带返回值的查询,(读)
// 返回一个[]map[string]interface 对应多行键值对
func (p *DbDriverMysql) Query(sql string, args ...interface{}) (data []map[string]interface{}, err error) {
	// SELECT

	stmt, err := p.db.Prepare(sql)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return
	}

	l := len(columns)
	scanArgs := make([]interface{}, l)
	values := make([]interface{}, l)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	data = []map[string]interface{}{}
	for rows.Next() {
		st := map[string]interface{}{}
		e := rows.Scan(scanArgs...)

		if e != nil {
			err = e
			return
		}
		for i, col := range values {
			if col != nil {
				st[columns[i]] = col
			}
		}

		data = append(data, st)
	}

	return
}

// 执行不带返回的查询(写)
// 返回insertId,
func (p *DbDriverMysql) Exec(sql string, args ...interface{}) (affectCount int64, lastInsertId int64, err error) {
	affectCount = 0
	lastInsertId = 0

	stmt, err := p.db.Prepare(sql)
	if err != nil {
		return
	}


	defer stmt.Close()
	result, err := stmt.Exec(args...)

	if err != nil {
		return
	}

	_affectCount, _ := result.RowsAffected()
	_lastInsertId, _ := result.LastInsertId()

	affectCount = _affectCount
	lastInsertId = _lastInsertId

	return
}
