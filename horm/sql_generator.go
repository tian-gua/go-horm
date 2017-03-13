package horm

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

type ISqlGenerator interface {
	GenerateListSql(i interface{}, conditions ...string) (string, error) //生成查询多条记录sql
	GenerateFindByIdSql(i interface{}) (string, error)                   //生成根据id查询sql
	GenerateSaveSql(i interface{}) (string, error)                       //生成保存记录sql
	GenerateUpdateByIdSql(i interface{}) (string, error)                 //生成根据id更新记录sql
	GenerateDelByIdSql(i interface{}) (string, error)                    //生成根据Id删除sql
}

var sqlGenerator ISqlGenerator = nil

//设置sql生成器
func SetSqlGenerator(sg ISqlGenerator) {
	sqlGenerator = sg
}

type defaultSqlGenerator struct {
}

func (d *defaultSqlGenerator) GenerateListSql(i interface{}, conditions ...string) (string, error) {
	structValue, err := getStructValue(i)
	if err != nil {
		return "", fmt.Errorf("get struct value error:%s", err.Error())
	}
	if structValue.pkName == "" || structValue.pkValue == "" {
		return "", fmt.Errorf("id can not be empty")
	}
	fields := ""
	for k, _ := range structValue.fieldValueMap {
		fields += k + ","
	}
	fields = strings.TrimSuffix(fields, ",")
	where := ""
	sort := ""
	for _, condition := range conditions {
		if strings.Contains(condition, "=") {
			where += " " + condition
		} else if strings.Contains(condition, "desc") || strings.Contains(condition, "DESC") || strings.Contains(condition, "asc") || strings.Contains(condition, "ASC") {
			sort += condition + ","
		}
	}
	sort = strings.TrimSuffix(sort, ",")
	if where != "" {
		where = "WHERE" + where
	}
	if sort != "" {
		sort = "ORDER BY " + sort
	}
	s := fmt.Sprintf("SELECT %s FORM %s %s %s", fields, structValue.tableName, where, sort)
	color.Green("[horm]εε[%s]:\t%s", time.Now().Format("2006-01-02 15:04:05"), s)
	return s, nil
}

func (d *defaultSqlGenerator) GenerateFindByIdSql(i interface{}) (string, error) {
	structValue, err := getStructValue(i)
	if err != nil {
		return "", fmt.Errorf("get struct value error:%s", err.Error())
	}
	if structValue.pkName == "" || structValue.pkValue == "" {
		return "", fmt.Errorf("id can not be empty")
	}

	fields := ""
	for k, _ := range structValue.fieldValueMap {
		fields += k + ","
	}
	fields = strings.TrimSuffix(fields, ",")
	s := fmt.Sprintf("SELECT %s FORM %s WHERE %s = %s", fields, structValue.tableName, structValue.pkName, structValue.pkValue)
	color.Green("[horm]εε[%s]:\t%s", time.Now().Format("2006-01-02 15:04:05"), s)
	return s, nil
}

func (d *defaultSqlGenerator) GenerateSaveSql(i interface{}) (string, error) {
	structValue, err := getStructValue(i)
	if err != nil {
		return "", fmt.Errorf("get struct value error:%s", err.Error())
	}
	if structValue.pkName == "" || structValue.pkValue == "" {
		color.Red("there is no primary key")
	}
	if len(structValue.fieldValueMap) == 0 {
		return "", fmt.Errorf("there is no field need to insert or no exported field")
	}
	fileds := structValue.pkName
	values := ""
	if structValue.autoIncrease {
		values += "DEFAULT"
	} else {
		values += structValue.pkValue
	}
	for k, v := range structValue.fieldValueMap {
		fileds += "," + k
		values += "," + v
	}
	s := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", structValue.tableName, fileds, values)
	color.Green("[horm]εε[%s]:\t%s", time.Now().Format("2006-01-02 15:04:05"), s)
	return s, nil
}

func (d *defaultSqlGenerator) GenerateUpdateByIdSql(i interface{}) (string, error) {
	structValue, err := getStructValue(i)
	if err != nil {
		return "", fmt.Errorf("get struct value error:%s", err.Error())
	}
	if structValue.pkName == "" || structValue.pkValue == "" {
		return "", fmt.Errorf("id can not be empty")
	}
	if len(structValue.fieldValueMap) == 0 {
		return "", fmt.Errorf("there is no need to update or no exported field")
	}
	set := ""
	for k, _ := range structValue.fieldValueMap {
		set += k + " = " + structValue.fieldValueMap[k] + ", "
	}
	set = strings.TrimSuffix(set, ", ")
	s := "UPDATE " + structValue.tableName + " SET " + set + " WHERE " + structValue.pkName + " = " + structValue.pkValue
	color.Green("[horm]εε[%s]:\t%s", time.Now().Format("2006-01-02 15:04:05"), s)
	return s, nil
}

func (d *defaultSqlGenerator) GenerateDelByIdSql(i interface{}) (string, error) {
	structValue, err := getStructValue(i)
	if err != nil {
		return "", fmt.Errorf("get struct value error:%s", err.Error())
	}
	if structValue.pkName == "" || structValue.pkValue == "" {
		return "", fmt.Errorf("id can not be empty")
	}
	s := fmt.Sprintf("DELETE FORM %s WHERE %s = %s", structValue.tableName, structValue.pkName, structValue.pkValue)
	color.Green("[horm]εε[%s]:\t%s", time.Now().Format("2006-01-02 15:04:05"), s)
	return s, nil
}
