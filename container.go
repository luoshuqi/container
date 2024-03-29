// 一个轻量的依赖注入容器。
//对象只实例化一次。
package container

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var c = &Container{
	instance:  map[instanceKey]reflect.Value{},
	resolving: map[reflect.Type]struct{}{},
}

// 依赖注入容器，存放已创建过的对象。
type Container struct {
	instance  map[instanceKey]reflect.Value
	resolving map[reflect.Type]struct{}
	lock      sync.Mutex
}

type instanceKey struct {
	ty  reflect.Type
	tag string
}

// 手动指定类型 T 的实例，用于 interface 或者初始化需要额外操作的 struct。
func Provide[T any](v T) {
	ProvideTagged(v, "")
}

// 手动指定类型 T 标记为 tag 的实例，用于 interface 或者初始化需要额外操作的 struct。
func ProvideTagged[T any](v T, tag string) {
	key := instanceKey{ty: typeof[T](), tag: tag}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.instance[key] = reflect.ValueOf(v)
}

// 获取类型 T 的实例，如果已创建过，直接返回，
//否则创建一个实例，自动设置所有没有标记 inject:"-" tag 的字段。
func Query[T any]() T {
	return RawQuery(typeof[T](), "").Interface().(T)
}

// 获取类型 T 标记为 tag 的实例，如果已创建过，直接返回，
//否则创建一个实例，自动设置所有没有标记 inject:"-" tag 的字段。
func QueryTagged[T any](tag string) T {
	return RawQuery(typeof[T](), tag).Interface().(T)
}

func RawQuery(ty reflect.Type, tag string) reflect.Value {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.query(ty, tag, nil)
}

func (r *Container) query(ty reflect.Type, tag string, parentTy reflect.Type) reflect.Value {
	key := instanceKey{ty: ty, tag: tag}
	if v, exists := r.instance[key]; exists {
		return v
	}
	if tag != "" {
		panic(fmt.Sprintf("instance of type %v with tag \"%v\" not found", ty, tag))
	}

	originalTy := ty
	ty = deref(ty)
	if ty.Kind() != reflect.Struct {
		if parentTy == nil {
			panic(fmt.Sprintf("cannot instantiate %v: not a struct", ty))
		} else {
			panic(fmt.Sprintf("cannot instantiate %v: expected struct found type %v", parentTy, ty))
		}
	}

	if _, exists := r.resolving[ty]; exists {
		panic(fmt.Sprintf("cannot instantiate %v: circular dependency %v", parentTy, ty))
	}
	r.resolving[ty] = struct{}{}
	defer delete(r.resolving, ty)

	v := reflect.New(ty)
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)
		inject := f.Tag.Get("inject")
		if inject != "-" {
			if !f.IsExported() {
				panic(fmt.Sprintf("unexported field %v.%v, add `inject:\"-\"` tag to skip inject", ty.String(), f.Name))
			}
			v.Elem().Field(i).Set(r.query(f.Type, parseTag(inject), ty))
		}
	}

	if originalTy.Kind() == reflect.Struct {
		v = v.Elem()
	}
	r.instance[key] = v
	return v
}

func parseTag(s string) string {
	if s == "" {
		return ""
	}
	for _, split := range strings.Split(s, ";") {
		if strings.HasPrefix(split, "tag:") {
			return split[4:]
		}
	}
	return ""
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
