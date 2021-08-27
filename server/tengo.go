package server

import (
	"ForkServer/tengolibs"
	"io/ioutil"
	"log"
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
	mm.AddBuiltinModule("url", tengolibs.URL)
	mm.AddBuiltinModule("server", tengolibs.Server(w, r, httpClient(), filepath.Join(pthPlugs, plug), map[string]tengo.Object{
		"version":     &tengo.String{Value: Vers},
		"script":      &tengo.String{Value: script},
		"plugin":      &tengo.String{Value: plug},
		"path":        &tengo.String{Value: path},
		"base_url":    &tengo.String{Value: "http://" + r.Host + "/" + plug + "/"},
		"torrserve":   &tengo.String{Value: sets.Torrserve},
		"assert":      &tengo.UserFunction{Name: "assert", Value: tengoAssert},
		"log_error":   &tengo.UserFunction{Name: "log_error", Value: tengoLog(plug, Err)},
		"log_warning": &tengo.UserFunction{Name: "log_warning", Value: tengoLog(plug, Wrn)},
		"log_info":    &tengo.UserFunction{Name: "log_info", Value: tengoLog(plug, Inf)},
		"memory":      &tengo.UserFunction{Name: "memory", Value: tengoMemory(plug)},
	}))
	t.SetImports(mm)
	_, e := t.Run()
	check(e)

}
func tengoLog(plg string, log *log.Logger) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		vs := []interface{}{plg + ":"}
		for _, a := range args {
			v, _ := tengo.ToString(a)
			vs = append(vs, v)
		}
		log.Println(vs...)
		return nil, nil
	}
}
func tengoAssert(args ...tengo.Object) (r tengo.Object, e error) {
	if len(args) != 2 {
		e = tengo.ErrWrongNumArguments
	} else if !args[0].IsFalsy() {
		r = tengo.TrueValue
	} else if i, o := args[1].(*tengo.Int); o {
		panic(i.Value)
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
