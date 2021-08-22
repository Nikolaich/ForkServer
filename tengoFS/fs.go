package tengofs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/d5/tengo/v2"
)

type fs struct {
	pth      string
	dwm, fwm os.FileMode
}

func FS(path string, permDir, permFile os.FileMode) map[string]tengo.Object {
	f := &fs{path, permDir, permFile}
	return map[string]tengo.Object{
		"inf":   &tengo.UserFunction{Name: "inf", Value: f.inf},
		"read":  &tengo.UserFunction{Name: "read", Value: f.read},
		"write": &tengo.UserFunction{Name: "write", Value: f.write},
	}
}
func (f *fs) path(args ...tengo.Object) (p string, e error) {
	if len(args) != 1 {
		e = tengo.ErrWrongNumArguments
	} else if a, o := args[0].(*tengo.String); o {
		p = filepath.Join(f.pth, filepath.Clean(a.Value))
	} else {
		e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string", Found: args[0].TypeName()}
	}
	return
}
func fi2map(i os.FileInfo) *tengo.Map {
	r := &tengo.Map{Value: map[string]tengo.Object{
		"name": &tengo.String{Value: i.Name()},
		"size": &tengo.Int{Value: i.Size()},
		"time": &tengo.Time{Value: i.ModTime()},
	}}
	if i.IsDir() {
		r.Value["dir"] = tengo.TrueValue
	} else {
		r.Value["dir"] = tengo.FalseValue
	}
	return r
}
func (f *fs) inf(args ...tengo.Object) (tengo.Object, error) {
	if p, e := f.path(args...); e != nil {
		return nil, e
	} else if i, e := os.Stat(p); e == nil {
		return fi2map(i), nil
	} else if os.IsNotExist(e) {
		return nil, nil
	} else {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	}
}
func (f *fs) read(args ...tengo.Object) (tengo.Object, error) {
	if p, e := f.path(args...); e != nil {
		return nil, e
	} else if i, e := os.Stat(p); os.IsNotExist(e) {
		return nil, nil
	} else if e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	} else if !i.IsDir() {
		if b, e := ioutil.ReadFile(p); e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		} else {
			return &tengo.Bytes{Value: b}, nil
		}
	} else if is, e := ioutil.ReadDir(p); e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	} else {
		r := new(tengo.Array)
		for _, i := range is {
			r.Value = append(r.Value, fi2map(i))
		}
		return r, nil
	}
}
func (f *fs) write(args ...tengo.Object) (tengo.Object, error) {
	a := false
	switch len(args) {
	case 3:
		a = !args[2].IsFalsy()
		fallthrough
	case 2:
		p, e := f.path(args[0])
		if e != nil {
			return nil, e
		} else if _, o := args[1].(*tengo.Undefined); o {
			var i os.FileInfo
			if i, e = os.Stat(p); os.IsExist(e) {
				return tengo.FalseValue, nil
			} else if e == nil {
				if i.IsDir() {
					if f.dwm == 0 {
						return nil, tengo.ErrNotImplemented
					} else {
						e = os.RemoveAll(p)
					}
				} else if f.fwm == 0 {
					return nil, tengo.ErrNotImplemented
				} else {
					e = os.Remove(p)
				}
			}
		} else if f.fwm == 0 {
			return nil, tengo.ErrNotImplemented
		} else {
			var fl *os.File
			if a {
				fl, e = os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, f.fwm)
			} else {
				fl, e = os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.fwm)
			}
			if e == nil {
				if b, o := args[1].(*tengo.Bytes); o {
					_, e = fl.Write(b.Value)
				} else {
					s, _ := tengo.ToString(args[1])
					_, e = fl.WriteString(s)
				}
			}
		}
		if e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		}
		return tengo.TrueValue, nil
	case 1:
		if f.dwm == 0 {
			return nil, tengo.ErrNotImplemented
		} else if p, e := f.path(args...); e != nil {
			return nil, e
		} else if e = os.Mkdir(p, f.dwm); os.IsExist(e) {
			return tengo.FalseValue, nil
		} else if e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		} else {
			return tengo.TrueValue, nil
		}
	default:
		return nil, tengo.ErrWrongNumArguments
	}
}
