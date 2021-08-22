package server

import (
	tengohttp "ForkServer/TengoHTTP"
	tengofs "ForkServer/tengoFS"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/d5/tengo/v2"
	stdlib "github.com/d5/tengo/v2/stdlib"
)

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
	mm.AddBuiltinModule("files", tengofs.FS(filepath.Join(pthPlugs, plug), 0777, 0666))
	mm.AddBuiltinModule("url", tengohttp.URL)
	mm.AddBuiltinModule("http", tengohttp.Client(httpClient()))
	mm.AddBuiltinModule("server", tengohttp.Server(w, r, map[string]tengo.Object{
		"version":     &tengo.String{Value: Vers},
		"script":      &tengo.String{Value: script},
		"plugin":      &tengo.String{Value: plug},
		"path":        &tengo.String{Value: path},
		"base_url":    &tengo.String{Value: "http://" + r.Host + "/" + plug + "/"},
		"torrserve":   &tengo.String{Value: sets.Torrserve},
		"assert":      &tengo.UserFunction{Name: "assert", Value: tengoAssert},
		"log_error":   &tengo.UserFunction{Name: "log_error", Value: tengoLog(plug, Error)},
		"log_warning": &tengo.UserFunction{Name: "log_warning", Value: tengoLog(plug, Warning)},
		"log_info":    &tengo.UserFunction{Name: "log_info", Value: tengoLog(plug, Info)},
	}))
	t.SetImports(mm)
	_, e := t.Run()
	check(e)

}
func tengoLog(plg string, log func(...interface{})) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		vs := []interface{}{plg}
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
		panic(i.Value)
	} else {
		s, _ := tengo.ToString(args[1])
		panic(s)
	}
	return
}
