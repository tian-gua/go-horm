package horm

import (
	"database/sql"
	"errors"
	"github.com/fatih/color"
	"strconv"
	"time"
)

type IHormManager interface {
	Connect(url string, port int, userName string, passWord string, dbName string) (int64, error) //连接数据库
	Create(int64) IHorm                                                                           //创建horm
	CloseAll() error                                                                              //关闭数据库连接
}

type HormManager struct {
	dbMap    map[int64]*sql.DB
	hormList []IHorm
}

func (m *HormManager) Connect(url string, port int, userName string, passWord string, dbName string) (int64, error) {
	db, err := sql.Open(MYSQL, userName+":"+passWord+"@tcp("+url+":"+strconv.Itoa(port)+")/"+dbName)
	if err != nil {
		panic(errors.New("Not connected to the database:" + err.Error()))
	}
	did := time.Now().UnixNano()
	m.dbMap[did] = db
	return did, nil
}

func (m *HormManager) Create(did int64) IHorm {
	return newDefaultHorm(m.dbMap[did])
}

func (m *HormManager) CloseAll() error {
	for k, v := range m.dbMap {
		err := v.Close()
		if err != nil {
			return errors.New("Connection closed failed:" + err.Error())
		}
		color.Green("[horm]εε[%s]:\tHorm-Connection[%d] is closed.\n", time.Now().Format("2006-01-02 15:04:05"), k)
	}
	return nil
}

//创建一个新的horm管理器
func New() IHormManager {
	return &HormManager{dbMap: make(map[int64]*sql.DB)}
}

//创建默认的Horm
func newDefaultHorm(db *sql.DB) IHorm {
	return &defaultHorm{db: db, mappings: newResultMap(), txMap: make(map[uint64]*sql.Tx)}
}
