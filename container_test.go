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
	Provide[fmt.Stringer](bar)
	foo := Query[*Foo]()
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
	Query[*Foo]()
}

func TestNonStruct(t *testing.T) {
	defer func() {
		if e := recover(); !matchMsg(e, "cannot instantiate int: not a struct") {
			t.Fail()
		}
	}()
	Query[int]()
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

	foo := Query[*Foo]()
	if foo.Bar != foo.Baz.Bar {
		t.Fail()
	}
}

func matchMsg(e any, msg string) bool {
	if s, ok := e.(string); ok && s == msg {
		return true
	}
	return false
}
