package server

import (
	tengohttp "ForkServer/tengoHTTP"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/d5/tengo/v2"
	stdlib "github.com/d5/tengo/v2/stdlib"
)

var tengomemory = make(map[string]tengo.Object)

func tengoRun(w http.ResponseWriter, r *http.Request, script, plug, path string) {
	var t *tengo.Script
	if b, e := ioutil.ReadFile(script); os.IsNotExist(e) {
		panic(404)
	} else if e != nil {
		panic(e)
	} else {
		t = tengo.NewScript(b)
	}
	t.EnableFileImport(true)
	t.SetImportDir(plug)
	mm := stdlib.GetModuleMap("math", "text", "times", "rand", "json", "base64", "hex")
	mm.AddBuiltinModule("server", tengohttp.ModuleMAP(w, r, httpClient(), map[string]tengo.Object{
		"version":   &tengo.String{Value: Vers},
		"script":    &tengo.String{Value: script},
		"plugin":    &tengo.String{Value: plug},
		"path":      &tengo.String{Value: path},
		"base_url":  &tengo.String{Value: "http://" + r.Host + "/" + plug + "/"},
		"torrserve": &tengo.String{Value: sets.Torrserve},
		"memory":    &tengo.UserFunction{Name: "memory", Value: tengoMemory(plug)},
		"file":      &tengo.UserFunction{Name: "file", Value: tengoFile(filepath.Join(pthPlugs, plug))},
		"assert":    &tengo.UserFunction{Name: "assert", Value: tengoAssert},
		"log_err":   &tengo.UserFunction{Name: "log_err", Value: tengoLog(plug, Error)},
		"log_wrn":   &tengo.UserFunction{Name: "log_inf", Value: tengoLog(plug, Warning)},
		"log_inf":   &tengo.UserFunction{Name: "log_inf", Value: tengoLog(plug, Info)},
	}))
	t.SetImports(mm)
	_, e := t.Run()
	check(e)

}
func tengoLog(plg string, log func(...interface{})) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		vs := []interface{}{plg + ":"}
		for _, a := range args {
			v, _ := tengo.ToString(a)
			vs = append(vs, v)
		}
		log(vs...)
		return nil, nil
	}
}
func tengoAssert(args ...tengo.Object) (r tengo.Object, e error) {
	if len(args) != 2 {
		e = tengo.ErrWrongNumArguments
	} else if !args[0].IsFalsy() {
		r = tengo.TrueValue
	} else if i, o := args[1].(*tengo.Int); o {
		panic(int(i.Value))
	} else {
		s, _ := tengo.ToString(args[1])
		panic(s)
	}
	return
}
func tengoMemory(p string) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if l := len(args); l > 1 {
			return nil, tengo.ErrWrongNumArguments
		} else if l == 0 {
			return tengomemory[p], nil
		} else if _, u := args[0].(*tengo.Undefined); u {
			delete(tengomemory, p)
			return nil, nil
		} else {
			tengomemory[p] = args[0]
			return tengomemory[p], nil
		}
	}
}
func tengoFile(pth string) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		var (
			p string
			e error
		)
		l := len(args)
		if l == 0 {
			return nil, tengo.ErrWrongNumArguments
		} else if n, o := args[0].(*tengo.String); !o {
			return nil, tengo.ErrInvalidArgumentType{Name: "first", Expected: "string", Found: args[0].TypeName()}
		} else {
			p = filepath.Join(pth, filepath.Clean(n.Value))
		}
		if l > 2 {
			var f *os.File
			prm := os.O_CREATE | os.O_WRONLY | os.O_APPEND
			if args[1].IsFalsy() {
				prm = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
			}
			if f, e = os.OpenFile(p, prm, 0666); e == nil {
				defer f.Close()
				for _, a := range args[2:] {
					if b, o := a.(*tengo.Bytes); o {
						f.Write(b.Value)
					} else if b, _ := tengo.ToString(a); b != "" {
						f.WriteString(b)
					}
				}
			}
		} else {
			args = append(args, tengo.FalseValue)
			switch v := args[1].(type) {
			case *tengo.Undefined:
				e = os.Remove(p)
			case *tengo.Bool:
				if v.IsFalsy() {
					var b []byte
					if b, e = ioutil.ReadFile(p); e == nil {
						return &tengo.Bytes{Value: b}, nil
					}
				} else {
					var i os.FileInfo
					if i, e = os.Stat(p); e == nil {
						r := &tengo.Map{Value: map[string]tengo.Object{
							"name": &tengo.String{Value: i.Name()},
							"size": &tengo.Int{Value: i.Size()},
							"time": &tengo.Time{Value: i.ModTime()},
						}}
						if i.IsDir() {
							r.Value["is_dir"] = tengo.TrueValue
						} else {
							r.Value["is_dir"] = tengo.FalseValue
						}
						return r, nil
					}
				}
			default:
				return nil, tengo.ErrInvalidArgumentType{Name: "second", Expected: "bool/undefined", Found: args[1].TypeName()}
			}
		}
		if os.IsNotExist(e) {
			return nil, nil
		} else if e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		}
		return tengo.TrueValue, nil
	}
}
