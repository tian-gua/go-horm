package horm

type Result struct {
	LastInsertId   int
	LastInsertId64 int64
	RowsAffected   int
	RowsAffected64 int64
}
