package horm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

//结构体信息
type StructInfo struct {
	tableName      string                          //表名
	structFieldMap map[string]*reflect.StructField //字段名->字段反射信息
	columnFieldMap map[string]string               //列名->字段名
	pkField        *reflect.StructField            //主键
}

//结构体字段值
type structValue struct {
	value          *reflect.Value            //结构体的值
	tableName      string                    //表名
	fieldStringMap map[string]string         //列名->字符串值
	fieldValueMap  map[string]*reflect.Value //列名->反射值
	pkStringValue  string                    //主键字符串值
	pkColumnName   string                    //主键的列名
	autoIncrease   bool                      //是否自增长
}

var structInfoMap map[string]*StructInfo

//获取结构体字段类型信息
func getStuctInfo(i interface{}) (*StructInfo, error) {
	t := reflect.TypeOf(i)
	/*校验参数是否是指针或者切片,如果是,则获取指向的元素的反射类型信息,如果参数不是结构体指针或者切片,返回错误*/
	kind := t.Kind()
	if kind == reflect.Ptr || kind == reflect.Slice {
		t = t.Elem()
		kind = t.Kind()
	}
	if kind != reflect.Struct {
		return nil, fmt.Errorf("[%s] is not struct", kind)
	}
	/*end*/
	/*从缓存中获反射信息*/
	if v, ok := structInfoMap[t.Name()]; ok {
		return v, nil
	}
	/*end*/
	var primarayKeyField *reflect.StructField = nil
	sfMap := make(map[string]*reflect.StructField)
	/*遍历结构体字段,保存带有field标签的字段类型信息*/
	for j := 0; j < t.NumField(); j++ {
		sf := t.Field(j)
		field := sf.Tag.Get(COLUMN_TAG)
		if field != "" {
			if field == "id" {
				primarayKeyField = &sf
			} else {
				sfMap[sf.Name] = &sf
			}
		}
	}
	/*end*/
	si := &StructInfo{structFieldMap: sfMap, pkField: primarayKeyField}
	/*通过table接口调用GetTableName方法获取表名*/
	if table, ok := i.(Table); ok {
		si.tableName = table.GetTableName()
	}
	/*end*/
	structInfoMap[kind.String()] = si //存放结构体类型信息到缓存里
	return si, nil
}

//获取结构体字段值
func getStructValue(i interface{}) (*structValue, error) {
	v := reflect.Indirect(reflect.ValueOf(i))
	sf, err := getStuctInfo(i) //获取参数类型反射信息
	if err != nil {
		return nil, fmt.Errorf("get struct info [%s] failed", v.Type().Name())
	}
	stringMap := make(map[string]string)
	valueMap := make(map[string]*reflect.Value)
	/*遍历结构体类型信息中保存的字段(过滤非field标签字段),并过滤非可导出的字段,获取字段的值*/
	for fieldName, structField := range sf.structFieldMap {
		value := v.FieldByName(fieldName)
		if !value.CanSet() {
			continue
		}
		convertedValue, err := convertString(value, value.Kind())
		if err != nil {
			return nil, fmt.Errorf("convert [value=%s type=%s] error", value.Type(), value.Kind().String())
		}
		stringMap[structField.Tag.Get(COLUMN_TAG)] = convertedValue
		valueMap[structField.Tag.Get(COLUMN_TAG)] = &value
	}
	/*end*/
	sv := &structValue{value: &v, fieldValueMap: valueMap, fieldStringMap: stringMap, tableName: sf.tableName}
	/*获取主键字段的值,校验主键字段是否可导出和是否是自增*/
	if sf.pkField != nil {
		sv.pkColumnName = sf.pkField.Tag.Get(COLUMN_TAG) //设置主键的列名
		pkValue := v.FieldByName(sf.pkField.Name)        //获取主键的反射值
		valueMap[sf.pkField.Tag.Get(COLUMN_TAG)] = &pkValue
		if !v.FieldByName(sf.pkField.Name).CanSet() {
			return nil, fmt.Errorf("primary key is unexported")
		}
		pkStringValue, err := convertString(v.FieldByName(sf.pkField.Name), sf.pkField.Type.Kind())
		if err != nil {
			return nil, fmt.Errorf("convert id error:%s", err.Error())
		}
		sv.pkStringValue = pkStringValue
		if "auto" == sf.pkField.Tag.Get("default") {
			autoIncrease := true
			sv.autoIncrease = autoIncrease
		}
	}
	/*end*/
	return sv, nil
}

//获取切片的元素
func getSliceElem(list interface{}) (interface{}, error) {
	v := reflect.Indirect(reflect.ValueOf(list))
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("[%s] not a slice", v.Kind())
	}
	elementType := v.Type().Elem()
	return reflect.New(elementType).Interface(), nil
}

//转换反射值为字符串值
func convertString(v reflect.Value, k reflect.Kind) (string, error) {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.Itoa(int(v.Int())), nil
	case reflect.Float64:
		return floatToString(v.Float()), nil
	case reflect.String:
		return "'" + v.String() + "'", nil
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format("'2006-01-02 15:04:05'"), nil
		}
	}
	return "", fmt.Errorf("convert value to string error:not support type[%s]", v.Type().Name())
}

//通过反射设置一个字段的值
func setValue(v *reflect.Value, rb sql.RawBytes) error {
	k := v.Kind()
	switch k {
	case reflect.Int:
		intValue, err := strconv.Atoi(string(rb))
		if err != nil {
			return fmt.Errorf("Set [%s] value failed:%s", k.String(), err)
		}
		v.Set(reflect.ValueOf(intValue))
	case reflect.Int8:
		intValue, err := strconv.Atoi(string(rb))
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int8(intValue)))
	case reflect.Int16:
		intValue, err := strconv.Atoi(string(rb))
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int16(intValue)))
	case reflect.Int32:
		intValue, err := strconv.Atoi(string(rb))
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int32(intValue)))
	case reflect.Int64:
		intValue, err := strconv.Atoi(string(rb))
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(int64(intValue)))
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(string(rb), 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(floatValue))
	case reflect.String:
		v.Set(reflect.ValueOf(string(rb)))
	case reflect.Struct:
		t, err := time.Parse("2006-01-02 15:04:05", string(rb))
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(t))
	}
	return nil
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
