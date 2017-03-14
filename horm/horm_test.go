package horm

import (
	"testing"
	"time"
)

func TestHorm(t *testing.T) {
	//创建一个HormManager
	hormManager := New()
	did, err := hormManager.Connect("127.0.0.1", 3306, "root", "root", "horm")
	dealError(err)
	t.Logf("did=%d", did)

	//创建一个horm
	horm := hormManager.Create(did)

	//开始一个事务
	err = horm.Begin()
	dealError(err)

	//创建一个测试struct
	th := newTestHorm()

	//保存新建的struct
	res, err := horm.Save(th)
	dealError(err)
	lastId, err := res.LastInsertId()
	dealError(err)

	//删除新建的struct
	th.Id = int(lastId)
	rows, err := horm.DelById(th)
	dealError(err)
	t.Logf("rows=%d", rows)

	//查询id = 9的struct
	th2 := &testHorm{Id: 9}
	err = horm.FindById(th2)
	dealError(err)
	t.Logf("%+v", th2)

	th2.Description = "更新 horm"
	rows, err = horm.UpdateById(th2)
	dealError(err)
	t.Logf("更新了[%d]条记录", rows)

	//查询id = 9的struct
	th2 = &testHorm{Id: 9}
	err = horm.FindById(th2)
	dealError(err)
	t.Logf("%+v", th2)

	ths := new([]testHorm)
	err = horm.List(ths)
	dealError(err)
	for _, v := range *ths {
		t.Logf("%+v", v)
	}

	//提交事务
	err = horm.Commit()
	dealError(err)

	//关闭所有连接
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
	State       int64     `field:"state"`
	Type        int       `field:"type"`
	Description string    `field:"description"`
}

func (t *testHorm) GetTableName() string {
	return "tb_test"
}

func newTestHorm() *testHorm {
	return &testHorm{CreateTime: time.Now(), ModifyTime: time.Now(), State: 0, Type: 0, Description: "测试horm"}
}
