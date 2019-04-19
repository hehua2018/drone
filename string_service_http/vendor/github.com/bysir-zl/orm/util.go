package orm

import (
	"reflect"
	"strings"
)

type FieldInfo struct {
	Name         string // 字段名
	Typ          reflect.Type
	CanInterface bool
	Tags         map[string]string
}

func DecodeStruct(prtStruct interface{}) map[string]FieldInfo {
	v := reflect.Indirect(reflect.ValueOf(prtStruct))
	t := v.Type()
	fieldNum := v.NumField()
	result := make(map[string]FieldInfo, fieldNum)
	for i := 0; i < fieldNum; i++ {
		field := t.Field(i)
		tags := EncodeTag(string(field.Tag))
		result[field.Name] = FieldInfo{
			Typ:         field.Type,
			CanInterface:v.Field(i).CanInterface(),
			Name:        field.Name,
			Tags:        tags,
		}
	}
	return result
}

func EncodeTag(tag string) (data map[string]string) {
	data = map[string]string{}
	if tag == "" {
		return
	}
	for _, item := range strings.Split(tag, " ") {
		if item == "" {
			continue
		}
		key := strings.Split(item, ":")[0]
		value := strings.Split(item, "\"")[1]
		data[key] = value
	}

	return
}

// 返回指定tag的 字段名=>tag value
func Field2TagMap(fieldInfo map[string]FieldInfo, tag string) map[string]string {
	result := map[string]string{}
	for _, info := range fieldInfo {
		for k, v := range info.Tags {
			if k == tag {
				result[info.Name] = v
			}
		}
	}
	return result
}

// 获取字段类型
func FieldType(fieldInfo map[string]FieldInfo) map[string]reflect.Type {
	result := map[string]reflect.Type{}
	for _, info := range fieldInfo {
		result[info.Name] = info.Typ
	}
	return result
}

func DecodeColumn(dbData string) map[string][]string {
	kvs:=map[string][]string{}
	if len(dbData)==0{
		return  kvs
	}

	ds := strings.Split(dbData, ";")
	l := len(ds)
	for i := 0; i < l; i++ {
		kv := ds[i]
		key := ""
		values := []string{""}
		if !strings.Contains(kv, "(") {
			key = kv
		} else {
			kAndV := strings.Split(kv, "(")
			key = kAndV[0]
			v := strings.Split(kAndV[1], ")")[0]
			values = strings.Split(v, ",")
		}
		kvs[key] = values
	}

	return kvs
}

// data


// 去重, 但是不能去除 int64 与 int 的重复 ...
func UnDuplicate(src []interface{}) []interface{} {
	if src == nil || len(src) == 0 {
		return src
	}
	temp := []interface{}{}
	has := map[interface{}]bool{}
	for i := range src {
		v := src[i]
		if _, ok := has[v]; !ok {
			has[v] = true
			temp = append(temp, v)
		}
	}

	return temp
}
