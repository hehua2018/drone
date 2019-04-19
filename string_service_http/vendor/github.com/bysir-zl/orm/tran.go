package orm

import (
	"encoding/json"
	"errors"
	"github.com/bysir-zl/bygo/util"
	"reflect"
	"strings"
	"time"
)

type Translator interface {
	Input(fieldName string, fieldType reflect.Type, input interface{}) (interface{}, error) // 入库
	Output(fieldName string, field reflect.Type, output interface{}) (interface{}, error)   // 出库
}

// jsonTran

type JsonTran struct{}

func (p *JsonTran) Input(fieldName string, fieldType reflect.Type, input interface{}) (result interface{}, err error) {
	bs, err := json.Marshal(input)
	if err != nil {
		return
	}
	result = util.B2S(bs)
	return
}
func (p *JsonTran) Output(fieldName string, fieldType reflect.Type, output interface{}) (result interface{}, err error) {
	s, ok := util.Interface2String(output, true)
	if !ok {
		err = errors.New(fieldName + " is't string, can't tran 'json'")
		return
	}
	ptrValue := reflect.New(fieldType)
	err = json.Unmarshal(util.S2B(s), ptrValue.Interface())
	if err != nil {
		return
	}
	result = ptrValue.Elem().Interface()

	return
}

type TimeTran struct{}

func (p *TimeTran) Input(fieldName string, fieldType reflect.Type, input interface{}) (result interface{}, err error) {
	if strings.Contains(fieldType.String(), "int") {
		// int => timeString
		s, _ := util.Interface2Int(input, true)
		t := time.Unix(s, 0).Format("2006-01-02 15:04:05")
		result = t
		return
	} else {
		// timeString => int
		s, ok := util.Interface2String(input, true)
		if !ok {
			err = errors.New(fieldName + " is't string, can't tran 'time'")
			return
		}
		t, e := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		if e != nil {
			err = e
			return
		}
		result = t.Unix()
		return
	}
}
func (p *TimeTran) Output(fieldName string, fieldType reflect.Type, output interface{}) (result interface{}, err error) {
	if strings.Contains(fieldType.String(), "int") {
		// 如果struct的字段是int型的,还要转换,则数据库里的是string型的
		// timeString  => int
		s, ok := util.Interface2String(output, true)
		if !ok {
			err = errors.New(fieldName + " is't string, can't tran 'time'")
			return
		}
		t, e := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		if e != nil {
			err  = e
			return
		}

		result = t.Unix()
		return
	} else {
		// int => timeString
		s, _ := util.Interface2Int(output, true)
		t := time.Unix(s, 0).Format("2006-01-02 15:04:05")
		result = t
		return
	}
}

//

func init() {
	RegisterTranslator("json", new(JsonTran))
	RegisterTranslator("time", new(TimeTran))
}
