# go-horm
轻量级的自动注册、多连接orm框架,操作方便

## 为什么写horm
因为工作的原因,会经常操作数据库,包括批量插入数据,表之间迁移,多表合并等操作,工作是用java,所以尝试过用java写,但是一些小操作用java写太浪费时间.
  
加上当时已经研究go一段时间了.所以尝试用go来写一个工具.然后就写了一个gorm(和jinzhu大大的gorm名字雷同..)
  
后来发现要做的事情越来越复杂,之前的orm写的很单纯,拓展性不好,所以这次写了一个接口化的orm框架
  
和其他orm框架一样,horm主要功能就是把struct和table关联起来.让用户对struct的操作映射到table上.

# 功能
+ 多连接(每一个goroutine可以拥有一个连接)
+ 自动注册
+ 包装常用操作(简单CRUD),自动生成sql语句
+ 支持事务(事务提交,事务回滚)
+ 接口化(包括horm管理器,horm对象,sql生成器都做了抽象),提供默认实现

### 支持的数据库
> 目前仅在mysql中测试过

#用法

### 测试用的struct
```
type testHorm struct {
	Id          int       `field:"id" default:"auto"`
	CreateTime  time.Time `field:"create_time"`
	ModifyTime  time.Time `field:"modify_time"`
	State       int       `field:"state"`
	Type        int       `field:"type"`
	Description string    `field:"description"`
}

//实现Table接口,用户自己指定表名
func (t *testHorm) GetTableName() string {
	return "tb_test"
}

func newTestHorm() *testHorm {
	return &testHorm{CreateTime:time.Now(), ModifyTime:time.Now(), State:0, Type:0, Description:"测试horm"}
}
```

### horm操作

```
//创建一个HormManager
hormManager := New()   
 
//连接数据库
//did为这一次连接的标识符,生成规则会当前时间戳(纳秒单位)
did, err := hormManager.Connect("127.0.0.1", 3306, "root", "root", "horm")   
   
//在当前did对应的连接中创建一个horm操作对象
horm := hormManager.Create(did)   
   
//开启一个事务
err = horm.Begin()
   
//保存新建的struct到数据库
//res为操作的结果,可以获取最新添加的id和操作的记录的条数
res, err := horm.Save(newTestHorm())
//提交事务
err = horm.Commit()    
   
//关闭所有goroutine的连接
err = hormManager.CloseAll()
```

### 控制台
```
[horm]εε[2017-03-13 11:31:15]:	INSERT INTO tb_test(id,description,create_time,modify_time,state,type) VALUES(DEFAULT,'测试horm','2017-03-13 11:31:15','2017-03-13 11:31:15',0,0)
[horm]εε[2017-03-13 11:31:15]:	Horm-Connection[1489375875531622312] is closed.
```

![horm使用步骤.png](https://github.com/aidonggua/go-horm/blob/master/horm%E4%BD%BF%E7%94%A8%E6%AD%A5%E9%AA%A4.png)