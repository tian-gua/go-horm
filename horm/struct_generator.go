package horm

import (
	"fmt"
	"strings"
)

type tableStruct struct {
	Field string `field:"Field"`
	Type  string `field:"Type"`
}

func GenerateStruct(h IHorm, tableName string, structName string) (string, error) {
	ts := new([]tableStruct)
	err := h.Query("DESC "+tableName, ts)
	if err != nil {
		return "", fmt.Errorf("Generate table struct failed:%s", err.Error())
	}
	structStr := fmt.Sprintf("\ntype %s struct {\n", structName)
	for _, v := range *ts {
		structStr += fmt.Sprintf("\t%-20s %-20s `field:\"%s\"`\n", getCamelString(v.Field), getStructType(v.Type), v.Field)
	}
	structStr += "}\n"
	structStr += fmt.Sprintf("func (%s *%s)GetTableName() string {\n\treturn \"%s\"\n}", ([]byte(structName))[0:1], structName, tableName)
	return structStr, nil
}

func getStructType(dbType string) string {
	if strings.Contains(dbType, "int") {
		return "int"
	} else if strings.Contains(dbType, "varchar") {
		return "string"
	} else if strings.Contains(dbType, "timestamp") || strings.Contains(dbType, "datetime") {
		return "time.Time"
	} else if strings.Contains(dbType, "decimal") || strings.Contains(dbType, "datetime") {
		return "float64"
	}
	return "unknown"
}

func getCamelString(unCamelString string) string {
	s := strings.Split(unCamelString, "_")
	camelString := ""
	for _, v := range s {
		camelString += strings.Title(v)
	}
	return camelString
}
