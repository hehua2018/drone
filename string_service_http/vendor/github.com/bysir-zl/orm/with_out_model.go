package orm

import (
	"errors"
	"strings"
	"time"
)

type WithOutModel struct {
	err     error
	connect string
	table   string
	fields  []string
	where map[string]([]interface{}) // condition => args
	order []orderItem
	limit [2]int
}

type orderItem struct {
	Field string
	Desc  string
}

func newWithOutModel() *WithOutModel {
	return &WithOutModel{
		connect: "default",
	}
}

func (p *WithOutModel) ExecSql(sql string, args ...interface{}) (affectCount int64, lastInsertId int64, err error) {
	c, err := config.writeConnect(p.connect)
	if err != nil {
		return
	}
	dbDriver, err := Singleton(c)
	if err != nil {
		return
	}
	t1 := time.Now()
	att, insertId, err := dbDriver.Exec(sql, args...)
	elapsed := time.Since(t1)
	info("SQL : "+sql, args, elapsed)
	if err != nil {
		return
	}

	lastInsertId = insertId
	affectCount = att

	return
}
func (p *WithOutModel) QuerySql(sql string, args ...interface{}) (result []map[string]interface{}, err error) {
	c, err := config.writeConnect(p.connect)
	if err != nil {
		return
	}
	dbDriver, err := Singleton(c)
	if err != nil {
		return
	}

	t1 := time.Now()
	result, err = dbDriver.Query(sql, args...)
	elapsed := time.Since(t1)
	info("SQL : "+sql, args, elapsed)
	if err != nil {
		return
	}

	return
}

func (p *WithOutModel) Table(table string) *WithOutModel {
	p.table = table
	return p
}

func (p *WithOutModel) GetTable() string {
	return p.table
}

func (p *WithOutModel) Connect(connect string) *WithOutModel {
	p.connect = connect
	return p
}

func (p *WithOutModel) Fields(fields ...string) *WithOutModel {
	p.fields = fields
	return p
}

func (p *WithOutModel) Limit(offset, size int) *WithOutModel {
	p.limit = [2]int{offset, size}
	return p
}

func (p *WithOutModel) Where(condition string, args ...interface{}) *WithOutModel {
	if p.where == nil {
		p.where = map[string][]interface{}{}
	}
	p.where[condition] = args
	return p
}

func (p *WithOutModel) WhereIn(condition string, args ...interface{}) *WithOutModel {
	if args == nil || len(args) == 0 {
		return p
	}
	if p.where == nil {
		p.where = map[string][]interface{}{}
	}
	if !strings.Contains(condition, "(?)") {
		p.err = errors.New("WhereIn condition must contains '(?)'")
		return p
	}
	s := strings.Repeat(",?", len(args))
	condition = strings.Replace(condition, "(?)", "("+s[1:]+")", -1)
	p.where[condition] = args
	return p
}

func (p *WithOutModel) Order(field string, desc string) *WithOutModel {
	if p.order == nil {
		p.order = []orderItem{}
	}
	p.order = append(p.order, orderItem{Field: field, Desc: desc})
	return p
}

func (p *WithOutModel) Insert(saveData map[string]interface{}) (id int64, err error) {
	if p.err != nil {
		err = p.err
		return
	}

	if p.fields != nil {
		// 过滤指定的字段
		temp := map[string]interface{}{}
		for _, k := range p.fields {
			temp[k] = saveData[k]
		}
		saveData = temp
	}

	sql, args, err := buildInsertSql(p.table, saveData)
	if err != nil {
		return
	}
	_, id, err = p.ExecSql(sql, args...)
	if err != nil {
		return
	}
	return
}

func (p *WithOutModel) Delete() (affect int64, err error) {
	if p.err != nil {
		err = p.err
		return
	}
	if p.where == nil || len(p.where) == 0 {
		err = errors.New("no where condition when DELETE")
		return
	}

	sql, args, err := buildDeleteSql(p.table, p.where)
	if err != nil {
		return
	}
	affect, _, err = p.ExecSql(sql, args...)
	if err != nil {
		return
	}
	return
}

func (p *WithOutModel) Update(saveData map[string]interface{}) (count int64, err error) {
	if p.err != nil {
		err = p.err
		return
	}
	if p.where == nil || len(p.where) == 0 {
		err = errors.New("no where condition when UPDATE")
		return
	}

	if p.fields != nil && len(p.fields) != 0 {
		// 过滤指定的字段
		temp := map[string]interface{}{}
		for _, k := range p.fields {
			temp[k] = saveData[k]
		}
		saveData = temp
	}

	sql, args, err := buildUpdateSql(p.table, saveData, p.where)
	if err != nil {
		return
	}
	count, _, err = p.ExecSql(sql, args...)
	if err != nil {
		return
	}
	return
}

func (p *WithOutModel) Select() (result []map[string]interface{}, has bool, err error) {
	if p.err != nil {
		err = p.err
		return
	}

	sql, args, err := buildSelectSql(p.fields, p.table, p.where, p.order, p.limit)
	if err != nil {
		return
	}
	result, err = p.QuerySql(sql, args...)
	if err != nil {
		return
	}
	has = len(result) != 0
	return
}

func (p *WithOutModel) First() (result map[string]interface{}, has bool, err error) {
	if p.err != nil {
		err = p.err
		return
	}
	p.limit = [2]int{0, 1}
	maps, has, err := p.Select()
	if !has || err != nil {
		return
	}
	result = maps[0]
	has = true
	return
}
