package orm

// IGetAttr 访问器
type IGetAttr interface {
	GetAttr()
}

// ISetAttr 修改器
type ISetAttr interface {
	SetAttr()
}

// IBeforeQuery 查询前钩子
type IBeforeQuery interface {
	BeforeQuery(*DB) error
}

// IAfterQuery 查询后钩子
type IAfterQuery interface {
	AfterQuery(*DB) error
}

// IBeforeCreate 创建前钩子
type IBeforeCreate interface {
	BeforeCreate(*DB) error
}

// IAfterCreate 查询后钩子
type IAfterCreate interface {
	AfterCreate(*DB) error
}

// IBeforeUpdate 修改前钩子
type IBeforeUpdate interface {
	BeforeUpdate(*DB) error
}

// IAfterUpdate 修改后钩子
type IAfterUpdate interface {
	AfterUpdate(*DB) error
}

// IBeforeDelete 删除前钩子
type IBeforeDelete interface {
	BeforeDelete(*DB) error
}

// IAfterDelete 删除后钩子
type IAfterDelete interface {
	AfterDelete(*DB) error
}
