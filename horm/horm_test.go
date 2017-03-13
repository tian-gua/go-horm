package horm

import (
	"testing"
	"time"
)

func TestHorm(t *testing.T) {
	hormManager := New() //创建一个HormManager
	did, err := hormManager.Connect("127.0.0.1", 3306, "root", "root", "horm")
	dealError(err)
	t.Logf("did=%d", did)
	horm := hormManager.Create(did)
	err = horm.Begin()
	dealError(err)
	th := newTestHorm()
	res, err := horm.Save(th)
	dealError(err)
	lastId, err := res.LastInsertId()
	dealError(err)
	th.Id = int(lastId)
	rows, err := horm.DelById(th)
	dealError(err)
	t.Logf("rows=%d", rows)
	err = horm.Commit()
	dealError(err)
	err = hormManager.CloseAll()
	dealError(err)
}

func dealError(err error) {
	if err != nil {
		panic(err)
	}
}

type testHorm struct {
	Id          int       `field:"id" default:"auto"`
	CreateTime  time.Time `field:"create_time"`
	ModifyTime  time.Time `field:"modify_time"`
	State       int       `field:"state"`
	Type        int       `field:"type"`
	Description string    `field:"description"`
}

func (t *testHorm) GetTableName() string {
	return "tb_test"
}

func newTestHorm() *testHorm {
	return &testHorm{CreateTime: time.Now(), ModifyTime: time.Now(), State: 0, Type: 0, Description: "测试horm"}
}
