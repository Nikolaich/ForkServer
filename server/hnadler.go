package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type pluginfo struct{ Title, Icon, Default string }

const extTengo, pthManifest, pthWebUI = ".tengo", "manifest.json", "index.html"

var torrLastCheck time.Time

func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			Warning(e, "ON REQUEST:", r.RequestURI)
			switch v := e.(type) {
			case int:
				http.Error(w, http.StatusText(v), v)
			case string:
				http.Error(w, v, http.StatusInternalServerError)
			case error:
				http.Error(w, v.Error(), http.StatusInternalServerError)
			}
		}
	}()
	Info("REQUEST:", r.RequestURI)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Set("Server", Name+"/"+Vers)
	switch p := strings.SplitN(r.URL.Path[1:], "/", 2); len(p) {
	case 1:
		handleSvc(w, r, p[0])
	case 2:
		handlePlg(w, r, p[0], p[1])
	default:
		panic(400)
	}
}
func handleSvc(w http.ResponseWriter, r *http.Request, n string) {
	switch n {
	case "":
		http.ServeFile(w, r, pthWebUI)
	case "test":
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><h1>ForkPlayer DLNA Work!</h1><b>" + Name + " v. " + Vers + " with plugins v. " + sets.PlugsTag + "</b><p>Runs on: " + runtime.GOOS + "/" + runtime.GOARCH + "</p></html>"))
	case "test.json":
		sets.stata(w)
	case "set":
		sets.set(w, r)
	case "treeview":
		treeview(w, r)
	case "parserlink":
		parserlink(w, r)
	case "torrserve":
		torrserve(w, r)
	case "proxy.m3u8":
		proxyM3U8(w, r)
	case "update":
		update(w, r)
	case "restart":
		go Restart()
		w.Write([]byte("Restarted within a second!"))
	default:
		image(w, r.URL.Path[1:])
	}
}
func handlePlg(w http.ResponseWriter, r *http.Request, n, p string) {
	if n == "treeview" {
		files(w, r)
	} else if i, e := plugInfo(n); os.IsNotExist(e) {
		panic(404)
	} else if e != nil {
		panic(e)
	} else if pth := filepath.Join(pthPlugs, n, filepath.Clean(strings.TrimSuffix(p, "/"))); strings.ToLower(filepath.Ext(pth)) == extTengo {
		tengoRun(w, r, pth, n, p)
	} else if f, e := os.Stat(pth); e == nil && !f.IsDir() {
		http.ServeFile(w, r, pth)
	} else if e != nil && !os.IsNotExist(e) {
		panic(e)
	} else if f, e := os.Stat(pth + extTengo); e == nil && !f.IsDir() {
		tengoRun(w, r, pth+extTengo, n, p)
	} else if i.Default != "" {
		tengoRun(w, r, filepath.Join(pthPlugs, n, filepath.Clean(i.Default)), n, p)
	} else {
		panic(404)
	}
}
func plugInfo(n string) (p pluginfo, e error) {
	var f *os.File
	if f, e = os.Open(filepath.Join(pthPlugs, n, pthManifest)); e == nil {
		e = json.NewDecoder(f).Decode(&p)
		f.Close()
	}
	return
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
