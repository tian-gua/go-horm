package horm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Table interface {
	GetTableName() string
}

//结构体信息
type StructInfo struct {
	kind           string                          //类型名
	tableName      string                          //表名
	structFieldMap map[string]*reflect.StructField //字段字典集合
	pkField        *reflect.StructField            //主键
}

//结构体字段值
type structFieldValue struct {
	tableName     string            //表名
	kind          string            //字段类型
	fieldValueMap map[string]string //字段属性值
	pkValue       string            //主键值
	pkName        string            //主键的值
	autoIncrease  bool              //是否自增长
}

var structInfoMap map[string]*StructInfo

//获取结构体字段类型信息
func getStuctInfo(i interface{}) (*StructInfo, error) {
	t := reflect.TypeOf(i)
	kind := t.Kind()
	if kind == reflect.Ptr {
		t = t.Elem()
		kind = t.Kind()
	}
	if v, ok := structInfoMap[kind.String()]; ok {
		return v, nil
	}
	if kind != reflect.Struct {
		return nil, fmt.Errorf("[%s] is not struct", kind)
	}
	var primarayKeyField *reflect.StructField = nil
	sfMap := make(map[string]*reflect.StructField)
	for j := 0; j < t.NumField(); j++ {
		sf := t.Field(j)
		field := sf.Tag.Get("field")
		if field != "" || sf.Tag.Get("default") != "" {
			if field == "id" {
				primarayKeyField = &sf
			} else {
				sfMap[sf.Name] = &sf
			}
		}
	}
	si := &StructInfo{kind: t.Kind().String(), structFieldMap: sfMap, pkField: primarayKeyField}
	if table, ok := i.(Table); ok {
		si.tableName = table.GetTableName()
	}
	structInfoMap[kind.String()] = si
	return si, nil
}

//获取结构体字段值
func getStructValue(i interface{}) (*structFieldValue, error) {
	v := reflect.Indirect(reflect.ValueOf(i))
	kind := v.Type().Kind()
	sf, err := getStuctInfo(i)
	if err != nil {
		return nil, fmt.Errorf("get struct info [%s] failed", kind)
	}
	valueMap := make(map[string]string)
	for fieldName, structField := range sf.structFieldMap {
		value := v.FieldByName(fieldName)
		if !value.CanSet() {
			continue
		}
		convertedValue, err := convertString(value, value.Kind())
		if err != nil {
			return nil, fmt.Errorf("convert [value=%s type=%s] error", value.Type(), value.Kind().String())
		}
		valueMap[structField.Tag.Get("field")] = convertedValue
	}
	if !v.FieldByName(sf.pkField.Name).CanSet() {
		return nil, fmt.Errorf("primary key is unexported")
	}
	pkValue, err := convertString(v.FieldByName(sf.pkField.Name), sf.pkField.Type.Kind())
	if err != nil {
		return nil, fmt.Errorf("convert id error")
	}
	autoIncrease := false
	if "auto" == sf.pkField.Tag.Get("default") {
		autoIncrease = true
	}
	return &structFieldValue{kind: kind.String(), fieldValueMap: valueMap, pkName: sf.pkField.Tag.Get("field"), pkValue: pkValue, tableName: sf.tableName, autoIncrease: autoIncrease}, nil
}

func convertString(v reflect.Value, k reflect.Kind) (string, error) {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.Itoa(int(v.Int())), nil
	case reflect.String:
		return "'" + v.String() + "'", nil
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format("'2006-01-02 15:04:05'"), nil
		} else {
			return "", fmt.Errorf("type ")
		}
	}
	return "", fmt.Errorf("convert value to string error")
}
