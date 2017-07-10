package main

import (
	"bytes"
	"fmt"
	"go/constant"
	"go/types"
	"sort"
	"strings"
)

var constStr = map[constant.Kind]string{
	constant.Bool:    "bool",
	constant.String:  "string",
	constant.Int:     "int",
	constant.Float:   "float",
	constant.Complex: "complex",
}

func panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

// packagePath returns the full, non-vendored package path of the package.
func packagePath(pkg *types.Package) string {
	const vendor = "/vendor/"

	p := pkg.Path()
	i := strings.LastIndex(p, vendor)
	if i < 0 {
		return p
	}
	return p[i+len(vendor):]
}

func formatAPI(pkg *types.Package, info *types.Info, buff *bytes.Buffer) {
	p := &printer{
		pkgPath: packagePath(pkg),
	}

	for _, def := range info.Defs {
		if def == nil || !def.Exported() {
			continue
		}

		// Only consider objects at the package scope.
		if def.Parent() != nil && def.Parent() != pkg.Scope() {
			continue
		}

		p.formatObj(def)
	}

	p.write(buff)
}

// ignoreRecv determines if we should ignored a delcared method based on the receiver's
// type. For example, we should ignore any methods on interfaces since they're printed
// when navigating the interface. We should also ignore any exported methods on unexported
// structs.
func ignoreRecv(recv types.Type) bool {
	switch t := recv.(type) {
	case *types.Named:
		if _, ok := t.Underlying().(*types.Interface); ok {
			// Ignore all methods with an interface receiver.
			return true
		}
		// Determine if the underlying type is exported.
		return !t.Obj().Exported()
	case *types.Pointer:
		// recurse
		return ignoreRecv(t.Elem())
	}
	return false
}

// isRecvInterface determines if a receiver is an interface type. Methods on interfaces
// are populated at the top level package declarations, so we need to be able to filter
// them out.
func isRecvInterface(recv *types.Var) bool {
	named, ok := recv.Type().(*types.Named)
	if !ok {
		return false
	}
	_, ok = named.Underlying().(*types.Interface)
	return ok
}

type printer struct {
	// lines to be printed out to the user.
	lines []string
	// Package that's being printed. This lets us omit the package name for
	// types within the package. For example if we're in package "a" ensure
	// types are printed as "T" and not "a.T"
	pkgPath string
}

// formatObj processes the given object, adding lines to the printer that represent that
// object's public API signature.
func (p *printer) formatObj(o types.Object) {

	switch o := o.(type) {
	case *types.Const:
		p.printf("const %s %s", o.Name(), constStr[o.Val().Kind()])

	case *types.Var:
		if o.IsField() {
			// Field on a struct. Ignore since this is captured when walking the strut.
			return
		}
		p.printf("var %s %s", o.Name(), p.formatType(o.Type().Underlying()))

	case *types.TypeName:
		switch u := o.Type().Underlying().(type) {
		case *types.Struct:
			p.printf("type %s struct", o.Name())
			for i := 0; i < u.NumFields(); i++ {
				f := u.Field(i)
				if !f.Exported() {
					continue
				}
				p.printf("type %s struct, %s %s", o.Name(), f.Name(), p.formatType(f.Type()))
			}

		case *types.Interface:
			var methodNames []string
			for i := 0; i < u.NumMethods(); i++ {
				f := u.Method(i)
				if !f.Exported() {
					continue
				}
				// Per package documentation, Type() of a Func is always a Signature.
				sig := f.Type().(*types.Signature)
				p.printf("type %s interface, %s%s", o.Name(), f.Name(), p.formatSignature(sig))
				methodNames = append(methodNames, f.Name())
			}
			if len(methodNames) == 0 {
				p.printf("type %s interface {}", o.Name())
			} else {
				sort.Strings(methodNames)
				p.printf("type %s interface { %s }", o.Name(), strings.Join(methodNames, ", "))
			}
		default:
			p.printf("type %s %s", o.Name(), p.formatType(u))
		}

	case *types.Func:
		// Per package documentation, Type() of a Func is always a Signature.
		sig := o.Type().(*types.Signature)
		if rec := sig.Recv(); rec != nil {
			if ignoreRecv(rec.Type()) {
				// Receiver is an interface value or unexported.
				return
			}
			p.printf("method (%s) %s%s", p.formatType(rec.Type()), o.Name(), p.formatSignature(sig))
		} else {
			p.printf("func %s%s", o.Name(), p.formatSignature(sig))
		}

	default:
		panicf("unexpected type %T %s", o, o)
	}
}

func (p *printer) printf(format string, v ...interface{}) {
	p.lines = append(p.lines, fmt.Sprintf(format, v...))
}

// formatType compactly represents a types signature.
func (p *printer) formatType(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return t.String()
	case *types.Struct:
		return t.String()
	case *types.Named:
		// TODO: Handle paths with "vendor" in them.
		typeName := t.Obj()
		if typeName.Pkg() == nil || packagePath(typeName.Pkg()) == p.pkgPath {
			// builtin type without a package like "error" or a type in
			// the package we're currently printing.
			return typeName.Name()
		}
		return packagePath(typeName.Pkg()) + "." + typeName.Name()
	case *types.Pointer:
		return "*" + p.formatType(t.Elem())
	case *types.Slice:
		return "[]" + p.formatType(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), p.formatType(t.Elem()))
	case *types.Signature:
		// Since this isn't the top level definition, there wont be a receiver.
		return "func" + p.formatSignature(t)
	case *types.Interface:
		return t.String()
	case *types.Map:
		return "map[" + p.formatType(t.Key()) + "]" + p.formatType(t.Elem())
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + p.formatType(t.Elem())
		case types.SendOnly:
			return "chan <- " + p.formatType(t.Elem())
		case types.RecvOnly:
			return "<- chan " + p.formatType(t.Elem())
		}
	default:
		panicf("unexpected type %T %s", t, t)
	}
	return ""
}

// formatSignature formats the arguments and return values of a function
// without inspecting the receiver or name. Example results include:
//
//		(string, string) error
//		(string) (bool, error)
//		(string) (*url.URL, error)
//
func (p *printer) formatSignature(sig *types.Signature) string {
	sigStr := "("
	params := sig.Params()

	// Does this signature end with a "..." argument?
	variadic := sig.Variadic()

	if params != nil {
		for i := 0; i < params.Len(); i++ {
			if i != 0 {
				sigStr += ", "
			}
			if variadic && i == params.Len()-1 {
				// If function is variadic and this is the last parameter it MUST be
				// of type "*types.Slice", right?
				sigStr += "..." + p.formatType(params.At(i).Type().(*types.Slice).Elem())
			} else {
				sigStr += p.formatType(params.At(i).Type())
			}
		}
	}
	sigStr += ")"

	r := sig.Results()

	if r != nil {
		switch r.Len() {
		case 0:
		case 1:
			// Special case a single result tuple and don't surround with ().
			sigStr += " " + p.formatType(r.At(0).Type())
		default:
			sigStr += " ("
			for i := 0; i < r.Len(); i++ {
				if i != 0 {
					sigStr += ", "
				}
				sigStr += p.formatType(r.At(i).Type())
			}
			sigStr += ")"
		}
	}

	return sigStr
}

func (p *printer) write(buff *bytes.Buffer) {
	sort.Strings(p.lines)
	for _, line := range p.lines {
		buff.WriteString("pkg ")
		buff.WriteString(p.pkgPath)
		buff.WriteString(", ")
		buff.WriteString(line)
		buff.WriteRune('\n')
	}
}
