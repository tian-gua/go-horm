package horm

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
	"sync"
)

type IHorm interface {
	List(list interface{}, conditions ...string) error //查询列表
	FindById(i interface{}) error                      //根据id查找
	Save(i interface{}) (*Result, error)               //插入单个记录
	UpdateById(i interface{}) (*Result, error)         //根据id更新
	DelById(i interface{}) (*Result, error)            //根据id删除
	Query(string, interface{}) error                   //自定义sql
	Exec(string) (*Result, error)                      //自定义sql
	Begin() error                                      //开始事务
	Commit() error                                     //提交事务
	RollBack() error                                   //回滚
	RegistMapping(i interface{}) error                 //注册映射(目前为自动注册)
}

type defaultHorm struct {
	db       *sql.DB
	mappings *resultMap
	txMap    map[uint64]*sql.Tx
	mutex    sync.Mutex
}

func (d *defaultHorm) List(list interface{}, conditions ...string) error {
	ele, err := getSliceElem(list)
	if err != nil {
		return fmt.Errorf("get slice element failed -> %s", err.Error())
	}
	sqlStr, err := sqlGenerator.GenerateListSql(ele, conditions...)
	if err != nil {
		return fmt.Errorf("Generate sql error:%s", err.Error())
	}
	rows, stmt, err := d.query(sqlStr)
	if err != nil {
		return fmt.Errorf("Query select sql error:%s", err)
	}
	err = injectStructList(list, ele, rows)
	if err != nil {
		return fmt.Errorf("Data inject error:%s", err)
	}
	err = rows.Close()
	if err != nil {
		return fmt.Errorf("close rows failed -> %s", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("Close statement error:%s", err.Error())
	}
	return nil
}

func (d *defaultHorm) FindById(i interface{}) error {
	sqlStr, err := sqlGenerator.GenerateFindByIdSql(i)
	if err != nil {
		return fmt.Errorf("generate sql error:%s", err.Error())
	}
	rows, stmt, err := d.query(sqlStr)
	if err != nil {
		return fmt.Errorf("Query select sql error:%s", err)
	}
	err = injectOneStruct(i, rows)
	if err != nil {
		return fmt.Errorf("Data inject error:%s", err)
	}
	err = rows.Close()
	if err != nil {
		return fmt.Errorf("close rows failed -> %s", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("Close statement error:%s", err.Error())
	}
	return nil
}

func (d *defaultHorm) Save(i interface{}) (*Result, error) {
	sqlStr, err := sqlGenerator.GenerateSaveSql(i)
	if err != nil {
		return nil, fmt.Errorf("generate sql failed:%s", err.Error())
	}
	return d.exec(sqlStr)
}

func (d *defaultHorm) UpdateById(i interface{}) (*Result, error) {
	sqlStr, err := sqlGenerator.GenerateUpdateByIdSql(i)
	if err != nil {
		return nil, errors.New("Generate sql failed:" + err.Error())
	}
	return d.exec(sqlStr)
}

func (d *defaultHorm) DelById(i interface{}) (*Result, error) {
	sqlStr, err := sqlGenerator.GenerateDelByIdSql(i)
	if err != nil {
		return nil, errors.New("Generate sql failed:" + err.Error())
	}
	return d.exec(sqlStr)
}

func (d *defaultHorm) Query(s string, i interface{}) error {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	printLog(s)
	rows, stmt, err := d.query(s)
	if err != nil {
		return fmt.Errorf("Query select sql error:%s", err)
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String:
		err = injectOneField(i, rows)
	case reflect.Struct:
		err = injectOneStruct(i, rows)
	case reflect.Slice:
		ele, err := getSliceElem(i)
		if err != nil {
			return fmt.Errorf("get slice element failed -> %s", err.Error())
		}
		if reflect.TypeOf(ele).Elem().Kind() == reflect.Struct {
			err = injectStructList(i, ele, rows)
		} else {
			err = injectOneFieldList(i, ele, rows)
		}
		if err != nil {
			return fmt.Errorf("Inject one field failed:%s", err.Error())
		}
	}
	err = rows.Close()
	if err != nil {
		return fmt.Errorf("Close rows failed:%s", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("Close statement error:%s", err.Error())
	}
	return nil
}

func (d *defaultHorm) Exec(s string) (*Result, error) {
	return d.exec(s)
}

func (d *defaultHorm) exec(sqlStr string) (*Result, error) {
	stmt, err := d.getStatement(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("Get statement error:%s", err.Error())
	}
	result, err := stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("Execute sql error:%s", err.Error())
	}
	err = stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("Close statement error:%s", err.Error())
	}
	lastInsertId64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	rowsAffected64, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	r := &Result{
		LastInsertId:   int(lastInsertId64),
		LastInsertId64: lastInsertId64,
		RowsAffected:   int(rowsAffected64),
		RowsAffected64: rowsAffected64,
	}
	return r, nil
}

func (d *defaultHorm) query(sqlStr string) (*sql.Rows, *sql.Stmt, error) {
	stmt, err := d.getStatement(sqlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("get statement error:%s", err.Error())
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, nil, fmt.Errorf("execute sql error:%s", err.Error())
	}
	return rows, stmt, nil
}

//向单个字段注入数据
func injectOneField(i interface{}, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		fmt.Errorf("get columns error:%s", err.Error())
	}
	if len(columns) != 1 {
		fmt.Errorf("found [%d] column but 1", len(columns))
	}
	rowNum := 0
	for rows.Next() {
		rowNum++
		if rowNum > 1 {
			return errors.New("select one but found more")
		}
		err = rows.Scan(i)
		if err != nil {
			return err
		}
	}
	return nil
}

//向单个结构体注入数据
func injectOneStruct(i interface{}, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		fmt.Errorf("get columns error:%s", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scans := make([]interface{}, len(columns))
	for index, _ := range values {
		scans[index] = &values[index]
	}
	sv, err := getStructValue(i)
	if err != nil {
		return fmt.Errorf("get struct value failed:%s", err.Error())
	}
	rowNum := 0
	for rows.Next() {
		rowNum++
		if rowNum > 1 {
			return errors.New("select one but found more")
		}
		err = rows.Scan(scans...)
		if err != nil {
			return err
		}
		for k, v := range values {
			f := sv.fieldValueMap[columns[k]]
			if f != nil {
				err = setValue(f, v)
				if err != nil {
					return fmt.Errorf("set value failed -> %s", err)
				}
			}
		}
	}
	return nil
}

//向单个字段切片注入数据
func injectOneFieldList(list interface{}, ele interface{}, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("get columns error:%s", err.Error())
	}
	if len(columns) != 1 {
		return fmt.Errorf("found [%d] columns but one", len(columns))
	}
	listValue := reflect.ValueOf(list).Elem()
	for rows.Next() {
		err = rows.Scan(ele)
		if err != nil {
			return err
		}
		listValue.Set(reflect.Append(listValue, reflect.ValueOf(ele).Elem()))
	}
	return nil
}

//向结构体切片注入数据
func injectStructList(list interface{}, ele interface{}, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		fmt.Errorf("get columns error:%s", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scans := make([]interface{}, len(columns))
	for index, _ := range values {
		scans[index] = &values[index]
	}
	listValue := reflect.ValueOf(list).Elem()
	sv, err := getStructValue(ele)
	if err != nil {
		return fmt.Errorf("get slice [%s] element struct reflect type failed -> %s", listValue.Type().Name(), err.Error())
	}
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return err
		}
		for k, v := range values {
			f := sv.fieldValueMap[columns[k]]
			if f != nil {
				err = setValue(f, v)
				if err != nil {
					return fmt.Errorf("set value failed -> %s", err)
				}
			}
		}
		listValue.Set(reflect.Append(listValue, *sv.value))
	}
	return nil
}

func (d *defaultHorm) Begin() error {
	printLog("transaction begin↓↓")
	d.mutex.Lock()
	var err error
	tx, err := d.db.Begin()
	if err != nil {
		return errors.New("transaction error -> " + err.Error())
	}
	d.txMap[getGID()] = tx
	return nil
}

func (d *defaultHorm) Commit() error {
	printLog("transaction commit↑↑")
	defer d.mutex.Unlock()
	return d.txMap[getGID()].Commit()
}

func (d *defaultHorm) RollBack() error {
	printLog("transaction rollback↑↑")
	defer d.mutex.Unlock()
	return d.txMap[getGID()].Rollback()
}

func (d *defaultHorm) RegistMapping(i interface{}) error {
	return errors.New("Not yet supported")
}

func (d *defaultHorm) getStatement(s string) (*sql.Stmt, error) {
	s = strings.TrimSpace(s)
	if d.txMap[getGID()] == nil {
		stmt, err := d.db.Prepare(s)
		return stmt, err
	}
	return d.txMap[getGID()].Prepare(s)
}
