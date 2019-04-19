package orm

import (
	"errors"
	"fmt"
	"github.com/bysir-zl/bygo/log"
	"github.com/bysir-zl/bygo/util"
	"reflect"
	"strings"
	"time"
)

type WithModel struct {
	WithOutModel
	modelInfo ModelInfo

	link map[string]linkData // objFieldName => linkData

	preLinkData map[InOneSql]PreLink
}

func newWithModel(ptrModel interface{}) *WithModel {
	w := &WithModel{}

	typ := reflect.TypeOf(ptrModel).String()
	typ = strings.Replace(typ, "*", "", -1)
	typ = strings.Replace(typ, "[]", "", -1)
	mInfo, ok := modelInfo[typ]
	if !ok {
		w.err = errors.New("can't found model " + typ + ",forget register?")
	} else {
		w.modelInfo = mInfo
	}
	w.table = w.modelInfo.Table
	w.connect = w.modelInfo.ConnectName
	return w
}

func (p *WithModel) Table(table string) *WithModel {
	p.WithOutModel.Table(table)
	return p
}

func (p *WithModel) Connect(connect string) *WithModel {
	p.WithOutModel.Connect(connect)
	return p
}

func (p *WithModel) Fields(fields ...string) *WithModel {
	p.WithOutModel.Fields(fields...)
	return p
}

func (p *WithModel) FieldsByModel(fields ...string) *WithModel {
	temp := []string{}
	for _, f := range fields {
		if db, ok := p.modelInfo.FieldMap[f]; ok {
			temp = append(temp, db)
		}
	}

	p.WithOutModel.Fields(temp...)
	return p
}

func (p *WithModel) Where(condition string, args ...interface{}) *WithModel {
	p.WithOutModel.Where(condition, args...)
	return p
}

func (p *WithModel) WhereIn(condition string, args ...interface{}) *WithModel {
	p.WithOutModel.WhereIn(condition, args...)
	return p
}

func (p *WithModel) Limit(offset, size int) *WithModel {
	p.WithOutModel.Limit(offset, size)
	return p
}

func (p *WithModel) Order(field, desc string) *WithModel {
	p.WithOutModel.Order(field, desc)
	return p
}

func (p *WithModel) Insert(prtModel interface{}) (err error) {
	if p.err != nil {
		err = p.err
		return
	}

	fieldData := map[string]interface{}{}
	// 读取保存的键值对
	mapper := util.ObjToMap(prtModel, "")
	for k, v := range mapper {
		// 在插入的时候过滤空值
		if !util.IsEmptyValue(v) {
			fieldData[k] = v
		}
	}
	// 自动添加字段
	autoSet, err := p.GetAutoSetField("insert")
	if err != nil {
		return
	}
	if autoSet != nil && len(autoSet) != 0 {
		for k, v := range autoSet {
			fieldData[k] = v
		}
		// 将自动添加的字段附加到model里，方便返回
		util.MapToObj(prtModel, autoSet, "")
	}

	// 转换值
	p.tranSaveData(&fieldData)

	// mapToDb
	dbData := map[string]interface{}{}
	for k, v := range fieldData {
		dbKey, ok := p.modelInfo.FieldMap[k]
		if ok {
			dbData[dbKey] = v
		}
	}

	id, err := p.WithOutModel.
		Insert(dbData)
	if err != nil {
		return
	}

	// 设置主键
	if p.modelInfo.AutoPk != "" && id != 0 {
		util.MapToObj(prtModel, map[string]interface{}{
			p.modelInfo.AutoPk: id,
		}, "")
	}

	return
}

func (p *WithModel) Update(prtModel interface{}) (count int64, err error) {
	if p.err != nil {
		err = p.err
		return
	}

	// 读取保存的键值对
	fieldData := util.ObjToMap(prtModel, "")

	// 自动添加字段
	autoSet, err := p.GetAutoSetField("update")
	if err != nil {
		return
	}
	if autoSet != nil && len(autoSet) != 0 {
		for k, v := range autoSet {
			fieldData[k] = v
		}
		// 将自动添加的字段附加到model里，方便返回
		util.MapToObj(prtModel, autoSet, "")
	}

	// 转换值
	p.tranSaveData(&fieldData)

	// mapToDb
	dbData := map[string]interface{}{}
	for k, v := range fieldData {
		dbKey, ok := p.modelInfo.FieldMap[k]
		if ok {
			dbData[dbKey] = v
		}
	}

	count, err = p.WithOutModel.
		Update(dbData)

	return
}

func (p *WithModel) Select(ptrSliceModel interface{}) (has bool, err error) {
	if p.err != nil {
		err = p.err
		return
	}
	// 是数组还是一个对象
	isSlice := strings.Contains(reflect.TypeOf(ptrSliceModel).String(), "[")
	if !isSlice {
		p.WithOutModel.limit = [2]int{0, 1}
	}
	result, has, err := p.WithOutModel.
		Select()
	if err != nil || !has {
		return
	}
	p.FromDbData(isSlice, result, ptrSliceModel)
	return
}

// 将从db里取得的map赋值到model里
func (p *WithModel) FromDbData(isSlice bool, result []map[string]interface{}, ptrSliceModel interface{}) {
	col2Field := util.ReverseMap(p.modelInfo.FieldMap)
	if isSlice {
		structData := make([]map[string]interface{}, len(result))
		for i, re := range result {
			structItem := make(map[string]interface{}, len(re))
			for k, v := range re {
				// 字段映射
				if structField, ok := col2Field[k]; ok {
					structItem[structField] = v
				}
			}
			// 转换值
			p.tranStructData(&structItem)
			p.preLink(&structItem)
			structData[i] = structItem
		}

		p.doLinkMulti(&structData)
		errInfo := util.MapListToObjList(ptrSliceModel, structData, "")
		if errInfo != "" {
			warn("table("+p.table+")", "tran", errInfo)
		}
	} else {
		resultItem := result[0]
		structItem := make(map[string]interface{}, len(resultItem))
		for k, v := range resultItem {
			// 字段映射
			if structField, ok := col2Field[k]; ok {
				structItem[structField] = v
			}
		}
		// 转换值
		p.tranStructData(&structItem)
		p.doLink(&structItem)
		//info("t",structItem)
		_, errInfo := util.MapToObj(ptrSliceModel, structItem, "")
		if errInfo != "" {
			warn("table("+p.table+")", "tran", errInfo)
		}
	}
}

type linkData struct {
	ExtCondition string
	Column       []string
}

// 连接对象
// 作者不推荐使用,这只是一个实验功能,太复杂不灵活,性能没保障
func (p *WithModel) Link(field string, extCondition string, columns []string) *WithModel {
	if p.link == nil {
		p.link = map[string]linkData{}
	}
	p.link[field] = linkData{ExtCondition: extCondition, Column: columns}
	return p
}

// 取得在method操作时需要自动填充的字段与值
func (p *WithModel) GetAutoSetField(method string) (needSet map[string]interface{}, err error) {
	autoFields := p.modelInfo.AutoFields
	if len(autoFields) != 0 {
		needSet = map[string]interface{}{}
		for field, auto := range autoFields {
			if util.ItemInArray(method, strings.Split(auto.When, "|")) {
				if auto.Typ == "time" {
					// 判断类型
					if strings.Contains(p.modelInfo.FieldTyp[field].String(), "int") {
						needSet[field] = time.Now().Unix()
					} else {
						needSet[field] = time.Now().Format("2006-01-02 15:04:05")
					}
				}
			}
		}
	}
	return
}

type PreLink struct {
	Args   []interface{} // 参数
	Column []string      // 要查询的字段
	Model  interface{}   // 要查询的模型(一个struct)
	ArgKey string        // 要连接的字段
}

// 能组装成一个条sql的
type InOneSql struct {
	Table      string
	WhereField string
}

func (p *InOneSql) String() string {
	return p.Table + "|" + p.WhereField
}

// 准备link
func (p *WithModel) preLink(data *map[string]interface{}) {
	if p.link == nil || len(p.link) == 0 {
		return
	}

	if p.preLinkData == nil {
		p.preLinkData = map[InOneSql]PreLink{} // key => PreLink
	}

	// 要link的字段
	for field, linkData := range p.link {
		// 判断有无field
		typ, ok := p.modelInfo.FieldTyp[field]
		if !ok {
			err := fmt.Errorf("have't %s field when link", field)
			warn("table("+p.table+")", err)
			continue
		}
		// 判断有无link属性
		link, ok := p.modelInfo.Links[field]
		if !ok {
			err := fmt.Errorf("have't link tag, plase use tag `orm:\"link(RoleId,Id)\"` on %s", field)
			warn("table("+p.table+")", err)
			continue
		}
		// 检查在原来的data中有无要连接的键的值
		val, ok := (*data)[link.SelfKey]
		if !ok {
			err := fmt.Errorf("have't '%s' value to link", link.SelfKey)
			warn("table("+p.table+")", err)
			continue
		}

		linkPtrValue := reflect.New(typ)

		where := "`" + link.LinkKey + "` in (?) AND " + linkData.ExtCondition
		one := InOneSql{
			Table:      newWithModel(linkPtrValue.Interface()).GetTable(),
			WhereField: strings.Trim(where, "AND "),
		}
		args := []interface{}{}

		// 要连接的是否是一个slice
		if typ.Kind() == reflect.Slice {
			valValue := reflect.ValueOf(val)
			// 检查值是否是一个slice
			// 只有slice才能连接slice
			if valValue.Kind() != reflect.Slice {
				err := fmt.Errorf("'%s' value is not slice to link slice", link.SelfKey)
				warn("table("+p.table+")", err)
				continue
			}
			vl := valValue.Len()
			vs := make([]interface{}, vl)
			for i := 0; i < vl; i++ {
				vs[i] = valValue.Index(i).Interface()
			}
			args = append(args, vs...)
		} else {
			args = append(args, val)
		}

		pre := PreLink{}
		pre.Model = util.GetElemInterface(linkPtrValue)
		pre.ArgKey = link.LinkKey
		if o, ok := p.preLinkData[one]; !ok {
			pre.Column = linkData.Column
			pre.Args = args
			p.preLinkData[one] = pre
		} else {
			pre.Args = append(o.Args, args...)
			p.preLinkData[one] = pre
		}
	}
	return
}

//
type ResultKeyMap struct {
	OneSql string
	Value  string // 由于数据库读出来的值可能和存放link值类型不对应(在第一个orm时会转换类型), 这里就全部转换为string去对应
}

func (p *WithModel) doLinkMulti(data *[]map[string]interface{}) {
	linkResult := map[ResultKeyMap]map[string]interface{}{} // onesql => key => model

	// 查询数据库
	for oneSql, pre := range p.preLinkData {
		pre.Args = UnDuplicate(pre.Args)
		rs, _, err := newWithOutModel().
			Connect(p.connect).Table(oneSql.Table).
			WhereIn(oneSql.WhereField, pre.Args...).
			Select()
		if err != nil {
			info("err", err)
			continue
		}
		for i, l := 0, len(rs); i < l; i++ {
			r := rs[i]
			value, ok := r[pre.ArgKey]
			if ok {
				vString, _ := util.Interface2StringWithType(value, false)
				resultKeyMap := ResultKeyMap{
					OneSql: oneSql.String(),
					Value:  vString,
				}
				linkResult[resultKeyMap] = r
			}
		}
	}

	for index, item := range *data {
		// 要link的字段
		for field, linkData := range p.link {
			// 判断有无field
			typ, _ := p.modelInfo.FieldTyp[field]

			// 判断有无link属性
			link, _ := p.modelInfo.Links[field]

			// 检查在原来的data中有无要连接的键的值
			// log.Info("xx",data)
			val := item[link.SelfKey]

			linkPtrValue := reflect.New(typ)

			where := "`" + link.LinkKey + "` in (?) AND " + linkData.ExtCondition
			linkModel := newWithModel(linkPtrValue.Interface())
			one := InOneSql{
				Table:      linkModel.GetTable(),
				WhereField: strings.Trim(where, "AND "),
			}
			// 要连接的是否是一个slice
			has := false

			if typ.Kind() == reflect.Slice {
				valValue := reflect.ValueOf(val)
				//info( valValue.Kind())
				if valValue.Kind() != reflect.Slice {
					err := fmt.Errorf("'%s' value is not slice to link slice", link.SelfKey)
					warn("table("+p.table+")", err)
					continue
				}
				models := []map[string]interface{}{}
				for i, l := 0, valValue.Len(); i < l; i++ {
					linkValue := valValue.Index(i).Interface()
					vString, _ := util.Interface2StringWithType(linkValue, false)

					resultKeyMap := ResultKeyMap{
						OneSql: one.String(),
						Value:  vString,
					}
					if model, ok := linkResult[resultKeyMap]; ok {
						models = append(models, model)
						has = true
					}
				}

				linkModel.FromDbData(true, models, linkPtrValue.Interface())
			} else {
				vString, _ := util.Interface2StringWithType(val, false)
				resultKeyMap := ResultKeyMap{
					OneSql: one.String(),
					Value:  vString,
				}

				if model, ok := linkResult[resultKeyMap]; ok {
					linkModel.FromDbData(false, []map[string]interface{}{model}, linkPtrValue.Interface())
					has = true
				}
			}

			if has {
				(*data)[index][field] = linkPtrValue.Elem().Interface()
			}
		}
	}

}

// 连接对象
// todo 在需要多次link的时候, 优化查询相同表(where in)
func (p *WithModel) doLink(data *map[string]interface{}) {
	p.preLinkData = nil

	if p.link == nil || len(p.link) == 0 {
		return
	}

	log.Info("sbsb", p.preLinkData)

	// 要link的字段
	for field, linkData := range p.link {
		// 判断有无field
		typ, ok := p.modelInfo.FieldTyp[field]
		if !ok {
			err := fmt.Errorf("have't %s field when link", field)
			warn("table("+p.table+")", err)
			continue
		}
		// 判断有无link属性
		link, ok := p.modelInfo.Links[field]
		if !ok {
			err := fmt.Errorf("have't link tag, plase use tag `orm:\"link(RoleId,Id)\"` on %s", field)
			warn("table("+p.table+")", err)
			continue
		}

		linkPtrValue := reflect.New(typ)

		// 检查在原来的data中有无要连接的键的值
		val := (*data)[link.SelfKey]
		if val == nil {
			err := fmt.Errorf("have't '%s' value to link", link.SelfKey)
			warn("table("+p.table+")", err)
			continue
		}

		m := newWithModel(linkPtrValue.Interface()).Fields(linkData.Column...)
		// 要连接的是否是一个slice
		if typ.Kind() == reflect.Slice {
			valValue := reflect.ValueOf(val)
			//info( valValue.Kind())
			if valValue.Kind() != reflect.Slice {
				err := fmt.Errorf("'%s' value is not slice to link slice", link.SelfKey)
				warn("table("+p.table+")", err)
				continue
			}
			vl := valValue.Len()
			vs := make([]interface{}, vl)
			for i := 0; i < vl; i++ {
				vs[i] = valValue.Index(i).Interface()
			}

			m = m.WhereIn("`"+link.LinkKey+"` in (?)", vs...)
		} else {
			m = m.Where("`"+link.LinkKey+"` = ?", val)
		}

		has, err := m.Select(linkPtrValue.Interface())

		if err != nil {
			warn("table("+p.table+")", err)
			continue
		}
		if has {
			(*data)[field] = linkPtrValue.Elem().Interface()
		}
	}

	return
}

// 将db的值 转换为struct的值
func (p *WithModel) tranStructData(saveData *map[string]interface{}) {
	for field, t := range p.modelInfo.Trans {
		v, ok := (*saveData)[field]
		if !ok {
			continue
		}

		if traner, ok := translators[t.Typ]; ok {
			data, err := traner.Output(field, p.modelInfo.FieldTyp[field], v)
			//info("tran", v,data,t.Typ)

			if err != nil {
				warn("table("+p.table+")", "tran", err)
				continue
			}
			(*saveData)[field] = data
		} else {
			warn("table("+p.table+")", "tran", "haven't traner named '"+t.Typ+"', forget register it ?")
			continue
		}
	}
	return
}

// 将struct的值 转换为db的值
func (p *WithModel) tranSaveData(saveData *map[string]interface{}) {
	for field, t := range p.modelInfo.Trans {
		v, ok := (*saveData)[field]
		if !ok {
			continue
		}

		if traner, ok := translators[t.Typ]; ok {
			data, err := traner.Input(field, p.modelInfo.FieldTyp[field], v)
			if err != nil {
				warn("table("+p.table+")", "tran", err)
				continue
			}
			(*saveData)[field] = data
		} else {
			warn("table("+p.table+")", "tran", "haven't traner named '"+t.Typ+"', forget register it ?")
			continue
		}
	}
	return
}
