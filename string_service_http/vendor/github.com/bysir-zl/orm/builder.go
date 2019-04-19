package orm

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)



func buildSelectSql(fields []string, tableName string,
where map[string]([]interface{}), order []orderItem, limit [2]int ) (sql string, args []interface{}, err error) {
	args = []interface{}{}
	sql = "SELECT "

	//field
	fieldString := "*"
	if fields != nil && len(fields) != 0 {
		fieldString =  strings.Join(fields, ",")
	}

	sql = sql + fieldString + " "

	//table
	sql = sql + "FROM `" + tableName + "` "

	//where
	if where != nil {
		whereString, as := buildWhere(where)
		for _, a := range as {
			args = append(args, a)
		}

		sql = sql + "WHERE " + whereString + " "
	}

	//orderBy
	if order != nil {
		orderString := ""
		for _, value := range order {
			orderString = orderString + "," + value.Field + " " + value.Desc
		}
		orderString = orderString[1:]

		sql = sql + "ORDER BY " + orderString + " "
	}

	//limit
	if limit[0] != 0 || limit[1] != 0 {
		sql = sql + "LIMIT " + fmt.Sprintf("%d,%d", limit[0], limit[1]) + " "
	}

	err = nil
	return
}

func buildInsertSql(tableName string, saveData map[string]interface{}) (sql string, args []interface{}, err error) {
	if saveData==nil||len(saveData) == 0 {
		err = errors.New("no save data on INSERT")
		return
	}
	if tableName==""{
		err= errors.New("not set table name")
		return
	}

	args = []interface{}{}
	sql = "INSERT INTO " + tableName + " ("

	var fields bytes.Buffer
	var holder bytes.Buffer

	for key, value := range saveData {
		fields.WriteString(",`" + key + "`")
		holder.WriteString(",?")
		args = append(args, value)
	}

	fieldsStr := fields.String()[1:]
	holderStr := holder.String()[1:]

	sql = sql + fieldsStr + " ) VALUES ( " + holderStr + " )"

	return
}

func buildUpdateSql(tableName string, saveData map[string]interface{}, where map[string]([]interface{})) (sql string, args []interface{}, err error) {

	if len(saveData) == 0 {
		err = errors.New("no save data on INSERT")
		return
	}

	args = []interface{}{}
	sql = "UPDATE " + tableName + " SET "

	//value
	var fields bytes.Buffer

	for key, value := range saveData {
		fields.WriteString(",`" + key + "`=?")
		args = append(args, value)
	}

	fieldsStr := fields.String()[1:]
	sql = sql + fieldsStr + " "

	//where
	if where != nil {
		whereString, as := buildWhere(where)
		for _, a := range as {
			args = append(args, a)
		}
		sql = sql + "WHERE " + whereString + " "
	}

	return
}

func buildDeleteSql(tableName string, where map[string]([]interface{})) (sql string, args []interface{}, err error) {
	args = []interface{}{}
	sql = "DELETE FROM " + tableName + " "

	//where
	if where != nil {
		whereString, as := buildWhere(where)
		args = as
		sql = sql + "WHERE (" + whereString + ") "
	}
	err = nil
	return
}

func buildCountSql(tableName string, where map[string]([]interface{})) (sql string, args []interface{}, err error) {
	sql = "SELECT COUNT(*) as count FROM " + tableName + " "

	//where
	if where != nil {
		whereString, as := buildWhere(where)
		args = as
		sql = sql + "WHERE (" + whereString + ") "
	}

	err = nil
	return
}

// 生成where 条件
func buildWhere(where map[string]([]interface{})) (whereString string, args []interface{}) {
	if where != nil {
		args = []interface{}{}
		whereString = " "

		for key, values := range where {
			whereString = whereString + " AND ( " + key + " )"
			for _, value := range values {
				args = append(args, value)
			}
		}

		whereString = whereString[5:]
	}

	return
}