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
	err := h.Query("DESC " + tableName, ts)
	if err != nil {
		return "", fmt.Errorf("Generate table struct failed:%s", err.Error())
	}
	structStr := fmt.Sprintf("\n type %s struct{\n", structName)
	for _, v := range *ts {
		structStr += fmt.Sprintf("\t%-20s %-20s\n", v.Field, getStructType(v.Type))
	}
	structStr += "}"
	return structStr, nil
}

func getStructType(dbType string) string {
	if strings.Contains(dbType, "int") {
		return "int"
	} else if strings.Contains(dbType, "varchar") {
		return "string"
	} else if strings.Contains(dbType, "timestamp") || strings.Contains(dbType, "datetime") {
		return "time.Time"
	}
	return "unknown"
}