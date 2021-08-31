package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const pthSettings = "settings.json"

type settings struct{ PlugsTag, Torrserve string }

var (
	stime = time.Now()
	sets  = new(settings)
)

func (s *settings) get() error {
	defer mutex.Unlock()
	mutex.Lock()
	f, e := os.Open(pthSettings)
	if e == nil {
		if e = json.NewDecoder(f).Decode(s); e != nil {
			e = errors.New("Parsing " + pthSettings + " error: " + e.Error())
		}
		f.Close()
	} else if os.IsNotExist(e) {
		e = nil
	}
	return e
}
func (s *settings) put() error {
	defer mutex.Unlock()
	mutex.Lock()
	f, e := os.Create(pthSettings)
	if e == nil {
		j := json.NewEncoder(f)
		j.SetIndent("", "  ")
		if e = j.Encode(s); e != nil {
			e = errors.New("Encoding settings error: " + e.Error())
		}
		f.Close()
	}
	return e
}
func (s *settings) set(w http.ResponseWriter, r *http.Request) {
	pl, ir := &playlist{Cmd: "reload(0.1);"}, r.FormValue("w_lang") == "ru"
	if a := r.FormValue("Torrserve"); a != "" {
		if a = r.FormValue(a); strings.LastIndexByte(a, ':') == -1 {
			a += ":8090"
		}
		if _, e := checkTorr(a); e != nil {
			if ir {
				pl.Info = "ОШИБКА: "
			} else {
				pl.Info = "ERROR: "
			}
			pl.Info += e.Error()
			pl.Cmd = "stop();"
		} else {
			pl.Note, s.Torrserve = " OK", a
			check(s.put())
		}
	}
	pl.write(w)
}
func (s *settings) stata(w http.ResponseWriter) {
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)
	m := struct{ Alloc, TotalAlloc, Sys, NumGC uint64 }{mem.Alloc, mem.TotalAlloc, mem.Sys, uint64(mem.NumGC)}
	d, _ := os.Getwd()
	i := struct {
		Name, Version, OS, Arch, Up, WD string
		HasUpdates                      bool
		*settings
		Memory, Treeview, Plugins interface{}
	}{Name, Vers, runtime.GOOS, runtime.GOARCH, ((time.Since(stime) / time.Second) * time.Second).String(), d, instaNew, s, &m, nil, nil}
	if f, e := os.Stat(pthTree); e == nil {
		i.Treeview = f.IsDir()
	} else if !os.IsNotExist(e) {
		i.Treeview = e.Error()
	}
	if fs, e := ioutil.ReadDir(pthPlugs); e != nil {
		i.Plugins = e.Error()
	} else if len(fs) > 0 {
		ps := make(map[string]interface{})
		for _, f := range fs {
			if n := f.Name(); f.IsDir() {
				if p, e := plugInfo(n); e != nil {
					ps[n] = e.Error()
				} else {
					ps[n] = p
				}
			}
		}
		i.Plugins = ps
	}
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	j.Encode(&i)
}
