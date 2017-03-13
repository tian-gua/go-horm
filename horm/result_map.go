package horm

//用于存放所有的结构体和表的映射
type resultMap struct {
	tableMap map[string]string              //结构体名->表名
	fieldMap map[string](map[string]string) //结构体名->(结构体字段->表字段)
}

func newResultMap() *resultMap {
	tableMap := make(map[string]string)
	fieldMap := make(map[string](map[string]string))
	return &resultMap{tableMap: tableMap, fieldMap: fieldMap}
}
