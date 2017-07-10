package main

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		want string
	}{
		{
			name: "type alias",
			pkg:  `type A string`,
			want: `pkg p, type A string`,
		},
		{
			name: "empty struct",
			pkg:  `type ContextKey struct{}`,
			want: `pkg p, type ContextKey struct`,
		},
		{
			name: "type alias import",
			pkg: `import "net/http"

			type Client http.Client`,
			want: `
pkg p, type Client struct
pkg p, type Client struct, CheckRedirect func(*net/http.Request, []*net/http.Request) error
pkg p, type Client struct, Jar net/http.CookieJar
pkg p, type Client struct, Timeout time.Duration
pkg p, type Client struct, Transport net/http.RoundTripper
			`,
		},
		{
			name: "variable",
			pkg:  "var B = 88",
			want: `pkg p, var B int`,
		},
		{
			name: "constant",
			pkg:  "const C = `99`",
			want: `pkg p, const C string`,
		},
		{
			name: "func",
			pkg:  `func Hello(a string, b int) (e error) { return nil }`,
			want: "pkg p, func Hello(string, int) error",
		},
		{
			name: "method",
			pkg: `
type A struct{}
func (a *A) Hello() (int, int)
func (a A) Bye(s string)

type a struct{}
func (a *a) Hello()
			`,
			want: `
pkg p, method (*A) Hello() (int, int)
pkg p, method (A) Bye(string)
pkg p, type A struct	
			`,
		},
		{
			name: "empty interface",
			pkg:  "type Foo interface {}",
			want: `pkg p, type Foo interface {}`,
		},
		{
			name: "interface",
			pkg: `
type Foo interface {
	Read([]uint8) (int, error)
	Close() error
}
			`,
			want: `
pkg p, type Foo interface { Close, Read }
pkg p, type Foo interface, Close() error
pkg p, type Foo interface, Read([]uint8) (int, error)
			`,
		},

		{
			name: "struct with fields",
			pkg: `
type I interface {
	Bar() string
}

type Foo struct {
	A I
	B map[string]I
}
			`,
			want: `
pkg p, type Foo struct
pkg p, type Foo struct, A I
pkg p, type Foo struct, B map[string]I
pkg p, type I interface { Bar }
pkg p, type I interface, Bar() string
			`,
		},
		{
			name: "dot dot dot",
			pkg:  `func Foo(b int, a ...interface{})`,
			want: `
    		pkg p, func Foo(int, ...interface{})
			`,
		},
		{
			name: "labels",
			pkg: `
func Foo(b int, a ...interface{}){
	goto Error
Error:
}`,
			want: `
    		pkg p, func Foo(int, ...interface{})
			`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const pkgHeader = "package p\n"

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "pkg.go", pkgHeader+test.pkg, 0)
			if err != nil {
				t.Fatal(err)
			}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}
			conf := types.Config{

				Importer: importer.Default(),
			}
			pkg, err := conf.Check("p", fset, []*ast.File{f}, info)
			if err != nil {
				t.Fatal(err)
			}

			buff := new(bytes.Buffer)
			formatAPI(pkg, info, buff)

			want := strings.TrimSpace(test.want)
			got := strings.TrimSpace(buff.String())

			if got != want {
				t.Fatalf("wanted:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}
