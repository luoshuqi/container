package container

import (
	"fmt"
	"testing"
)

type Bar struct {
	s string
}

func (r *Bar) String() string {
	return r.s
}

func TestProvide(t *testing.T) {
	type Foo struct {
		S fmt.Stringer
	}
	bar := &Bar{s: "bar"}
	Provide[fmt.Stringer](bar, None)
	foo := Query[*Foo](None)
	if foo.S != bar {
		t.Fail()
	}
}

type Baz struct {
	Foo *Foo
}

type Foo struct {
	Baz *Baz
}

func TestCircularDependency(t *testing.T) {
	defer func() {
		if e := recover(); !matchMsg(e, "cannot instantiate container.Baz: circular dependency container.Foo") {
			t.Fail()
		}
	}()
	Query[*Foo](None)
}

func TestNonStruct(t *testing.T) {
	defer func() {
		if e := recover(); !matchMsg(e, "cannot instantiate int: not a struct") {
			t.Fail()
		}
	}()
	Query[int](None)
}

func TestSingleton(t *testing.T) {
	type Bar struct {
		s string
	}

	type Baz struct {
		s   string
		Bar *Bar
	}

	type Foo struct {
		Bar *Bar
		Baz *Baz
	}

	foo := Query[*Foo](None)
	if foo.Bar != foo.Baz.Bar {
		t.Fail()
	}
}

func TestIgnoreTag(t *testing.T) {
	type Foo struct {
		X int `inject:"-"`
	}
	Query[*Foo](None)
}

type CacheDriver interface {
	name() string
}

type FileCacheDriver struct{}

func (r *FileCacheDriver) name() string {
	return "file"
}

type RedisCacheDriver struct{}

func (r *RedisCacheDriver) name() string {
	return "redis"
}

func TestTag(t *testing.T) {
	type Foo struct {
		File  CacheDriver
		Redis CacheDriver `inject:"tag:redis"`
	}

	Provide[CacheDriver](Query[*FileCacheDriver](None), None)
	Provide[CacheDriver](Query[*RedisCacheDriver](None), "redis")

	f := Query[Foo](None)
	if f.File.name() != "file" {
		t.Fail()
	}
	if f.Redis.name() != "redis" {
		t.Fail()
	}
}

func matchMsg(e any, msg string) bool {
	if s, ok := e.(string); ok && s == msg {
		return true
	}
	return false
}

func TestExample(t *testing.T) {
	Example()
}
