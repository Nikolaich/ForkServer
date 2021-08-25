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
	stime    = time.Now()
	sets     = new(settings)
	shuffles = make(map[string]bool)
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
	i := struct {
		Name, Version, OS, Arch string
		Up                      time.Duration
		HasUpdates              bool
		*settings
		Treeview, Plugins interface{}
		MemStats          *runtime.MemStats
	}{Name, Vers, runtime.GOOS, runtime.GOARCH, time.Since(stime) / time.Second, instaNew, s, nil, nil, new(runtime.MemStats)}
	runtime.ReadMemStats(i.MemStats)
	if f, e := os.Stat(pthTree); os.IsNotExist(e) {
		i.Treeview = nil
	} else if e != nil {
		i.Treeview = e.Error()
	} else {
		i.Treeview = f.IsDir()
	}
	if fs, e := ioutil.ReadDir(pthPlugs); e != nil {
		i.Plugins = e.Error()
	} else {
		var ps []interface{}
		for _, f := range fs {
			if n := f.Name(); f.IsDir() {
				if p, e := plugInfo(n); e != nil {
					ps = append(ps, e.Error())
				} else {
					ps = append(ps, p)
				}
			}
		}
		i.Plugins = ps
	}
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	j.Encode(&i)
}
