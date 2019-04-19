# orm
bysir-zl/rom is a supported 
是一个golang编写支持的 存json与模拟join 的orm

## 使用 Usage

```
go get github.com/bysir-zl/orm
```
下面是一个demo+注解
```go
package tests
import (
	"github.com/bysir-zl/bygo/util"
	"github.com/bysir-zl/orm"
	"log"
	"testing"
)
type User struct {
	orm string `table:"user" connect:"default" json:"-"`

	Id         int    `orm:"col(id);pk(auto);" json:"id"`
	Name       string `orm:"col(name)" json:"name"`
	Sex        bool `orm:"col(sex)" json:"sex"`
	Role_ids   []int `orm:"col(role_ids);tran(json)" json:"role_ids"`
	RoleId     int `orm:"col(role_id)"  json:"stime"`
	Created_at string `orm:"col(created_at);auto(insert,time)"  json:"stime"`
	Updated_at string `orm:"col(updated_at);auto(insert|update,time);tran(time)" json:"itime"`
	RoleRaw *Role `orm:"col(role_raw);tran(json)"`
}

type Role struct {
	orm string `table:"role" connect:"default" json:"-"`

	Id   int    `orm:"col(id);pk(auto);" json:"id"`
	Name string `orm:"col(name)" json:"name"`
}

func TestInsert(t *testing.T) {
	test := User{
		Name:"bysir",
		RoleId:1,
		Role_ids:[]int{1,2,3},
		Sex:true,
		RoleRaw:&Role{
			Name:"inJson",
			Id:  1,
		},
	}
	// 你可能会注意到RoleRaw和Role_ids字段是一个结构体, 他们会自动在入库的时候序列化为json串,
	// 在出库的时候自动反序列化为结构体, 这只需要在你的model字段上添加一个 tran(json) Tag, 内置两个转换器是json和time,
	// 要自定义你只需要实现一个转换器(Translator)并RegisterTranslator(name,Translator)

	err := orm.Model(&test).Insert(&test)
	if err != nil {
		t.Error(err)    
	}
}

// 别忘了配置Db与注册模型
func init() {
    // 开启debug会打印sql语句
	orm.Debug = true

	orm.RegisterDb("default", "mysql", "root:@tcp(localhost:3306)/test")
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Role))
}
````

## 高级用法 Advanced

### link(like join)
实现模拟数据库join语法, 为什么不直接用join? 个人不喜欢而已 ;)

假如你要实现根据一个role_id关联到role表的数据的话,你可能会想到用join, 现在还有一个前端实现: link, orm会帮你自动的读取关联模型,

看看下面这个model
```go
type User struct {
	orm string `table:"user" connect:"default" json:"-"`

	Id         int    `orm:"col(id);pk(auto);" json:"id"`
	Name       string `orm:"col(name)" json:"name"`
	Role_ids   []int `orm:"col(role_ids);tran(json)" json:"role_ids"`
	RoleId     int `orm:"col(role_id)"  json:"stime"`

	Role   *Role `orm:"link(RoleId,id)"`
	Roles []Role `orm:"link(Role_ids,id)"`
}
```
这样使用link
```go
func TestSelect(t *testing.T) {
	us := []User{}
	
	// 需要显示的指定你要link的字段
	_, err := orm.Model(&us).
		Link("Role", "", []string{"name"}).
		Link("Roles", "", nil).
		Select(&us)
	if err != nil {
		t.Error(err)
	}
	log.Printf("role    %+v", ts[0].Role) // 一个role
	log.Printf("roles   %+v", ts[0].Roles)// role的slice
}
```
就是这么简单, 当查询多个Role的时候, orm还会自动优化sql, 解决sql n+1 问题