package horm

func init() {
	SetSqlGenerator(&defaultSqlGenerator{})
	structInfoMap = make(map[string]*StructInfo)
}
