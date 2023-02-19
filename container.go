// 一个轻量的依赖注入容器。
//
//对象只实例化一次，非并发安全。
//
// 示例：
//
//
package container

import (
	"fmt"
	"reflect"
)

var defaultContainer = NewContainer()

// 依赖注入容器，存放已创建过的对象。
type Container struct {
	instance  map[reflect.Type]reflect.Value
	resolving map[reflect.Type]struct{}
}

// 创建一个容器。
func NewContainer() *Container {
	return &Container{
		instance:  map[reflect.Type]reflect.Value{},
		resolving: map[reflect.Type]struct{}{},
	}
}

// 手动指定类型 T 的实例，用于 interface 或者初始化需要额外操作的 struct。
func ProvideWith[T any](v T, container *Container) {
	container.instance[typeof[T]()] = reflect.ValueOf(v)
}

// 类似 ProvideWith，使用默认容器。
func Provide[T any](v T) {
	ProvideWith(v, defaultContainer)
}

// 获取类型 T 的实例，如果已创建过，直接返回，
//
//否则创建一个实例，自动设置所有已导出并且没有 container:"-" tag 的字段。
func QueryWith[T any](container *Container) T {
	return query(container, typeof[T](), nil).Interface().(T)
}

// 类似 QueryWith，使用默认容器。
func Query[T any]() T {
	return QueryWith[T](defaultContainer)
}

func query(container *Container, t reflect.Type, st reflect.Type) reflect.Value {
	if v, exists := container.instance[t]; exists {
		return v
	}

	ot := t
	t = deref(t)
	if t.Kind() != reflect.Struct {
		if st == nil {
			panic(fmt.Sprintf("cannot instantiate %v: not a struct", t))
		} else {
			panic(fmt.Sprintf("cannot instantiate %v: expected type struct found type %v", st, t))
		}
	}

	if _, exists := container.resolving[t]; exists {
		panic(fmt.Sprintf("cannot instantiate %v: circular dependency %v", st, t))
	}
	container.resolving[t] = struct{}{}

	v := reflect.New(t)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.IsExported() && f.Tag.Get("container") != "-" {
			v.Elem().Field(i).Set(query(container, f.Type, t))
		}
	}

	if ot.Kind() == reflect.Struct {
		v = v.Elem()
	}
	container.instance[ot] = v
	delete(container.resolving, t)
	return v
}

func typeof[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}
	return t
}
