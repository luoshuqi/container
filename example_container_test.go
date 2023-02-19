package container

import (
	"fmt"
)

type UserManager struct {
	Cache   Cache
	Storage *StorageManger
}

type Cache interface {
	Get(key string) string
}

type FileCache struct {
	Storage *StorageManger
}

func (r *FileCache) Get(key string) string {
	return ""
}

type StorageManger struct{}

func Example() {
	// 指定 Cache 接口的实例
	Provide[Cache](Query[*FileCache]())
	// 获取 *UserManager 实例
	userManager := Query[*UserManager]()
	fmt.Println(userManager)
}
