package horm

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type IHorm interface {
	List(list []interface{}, conditions ...string) error     //查询列表
	FindById(i interface{}) error                            //根据id查找
	Save(i interface{}) (sql.Result, error)                  //插入单个记录
	Update(i interface{}, conditions ...string) (int, error) //根据条件更新
	UpdateById(i interface{}) (int64, error)                   //根据id更新
	DelById(i interface{}) (int64, error)                      //根据id删除
	Del(i interface{}, conditions ...string) (int, error)    //根据条件删除
	Query(string) (int, error)                               //自定义sql
	Begin() error                                            //开始事务
	Commit() error                                           //提交事务
	RollBack() error                                         //回滚
	RegistMapping(i interface{}) error                       //注册映射(目前为自动注册)
}

type defaultHorm struct {
	db       *sql.DB
	mappings *resultMap
	txMap    map[uint64]*sql.Tx
	mutex    sync.Mutex
}

func (d *defaultHorm) List(list []interface{}, conditions ...string) error {
	return errors.New("Not yet supported")
}

func (d *defaultHorm) FindById(i interface{}) error {
	return errors.New("Not yet supported")
}

func (d *defaultHorm) Save(i interface{}) (sql.Result, error) {
	sqlStr, err := sqlGenerator.GenerateSaveSql(i)
	if err != nil {
		return nil, fmt.Errorf("generate sql error:%s", err.Error())
	}
	return d.exec(sqlStr)
}

func (d *defaultHorm) Update(i interface{}, conditions ...string) (int, error) {
	return 0, errors.New("Not yet supported")
}

func (d *defaultHorm) UpdateById(i interface{}) (int64, error) {
	sqlStr, err := sqlGenerator.GenerateUpdateByIdSql(i)
	if err != nil {
		return 0, errors.New("Generate sql failed:" + err.Error())
	}
	res, err := d.exec(sqlStr)
	if err != nil {
		return 0, errors.New("Execute update operate failed:" + err.Error())
	}
	return res.RowsAffected()
}

func (d *defaultHorm) DelById(i interface{}) (int64, error) {
	sqlStr, err := sqlGenerator.GenerateDelByIdSql(i)
	if err != nil {
		return 0, errors.New("Generate sql failed:" + err.Error())
	}
	res, err := d.exec(sqlStr)
	if err != nil {
		return 0, errors.New("Execute delete operate failed:" + err.Error())
	}
	return res.RowsAffected()
}

func (d *defaultHorm) Del(i interface{}, conditions ...string) (int, error) {
	return 0, errors.New("Not yet supported")
}

func (d *defaultHorm) Query(s string) (int, error) {
	return 0, errors.New("Not yet supported")
}

func (d *defaultHorm) Begin() error {
	d.mutex.Lock()
	var err error
	tx, err := d.db.Begin()
	if err != nil {
		return errors.New("Begin transaction failed:" + err.Error())
	}
	d.txMap[getGID()] = tx
	return nil
}

func (d *defaultHorm) Commit() error {
	defer d.mutex.Unlock()
	return d.txMap[getGID()].Commit()
}

func (d *defaultHorm) RollBack() error {
	defer d.mutex.Unlock()
	return d.txMap[getGID()].Rollback()
}

func (d *defaultHorm) RegistMapping(i interface{}) error {
	return errors.New("Not yet supported")
}

func (d *defaultHorm) getStatement(s string) (*sql.Stmt, error) {
	if d.txMap[getGID()] == nil {
		stmt, err := d.db.Prepare(s)
		return stmt, err
	}
	return d.txMap[getGID()].Prepare(s)
}

func (d *defaultHorm)exec(sqlStr string) (sql.Result, error) {
	stmt, err := d.getStatement(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("get statement error:%s", err.Error())
	}
	result, err := stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("execute sql error:%s", err.Error())
	}
	return result, nil
}