package server

import (
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const pthTree, pthPlugs = "treeview", "plugins"

var fileTypes = map[byte][]string{
	't': {"torrent"},
	'l': {"m3u"},
	'a': {"mp3", "wav", "ogg", "wma"},
	'v': {"avi", "mp4", "mkv", "ts", "mpeg", "mpg", "mov", "webm", "wmv", "mts", "vob", "3gp", "3g2", "trp", "tp", "dat", "asf"},
}

func treeview(w http.ResponseWriter, r *http.Request) {
	pl, ir := &playlist{r: r, Cache: "nocache", Type: "start"}, r.FormValue("w_lang") == "ru"
	if instaNew {
		t := "Install updates!"
		if ir {
			t = "Установить обновления!"
		}
		pl.add("/update", t, "/oyd.svg", false)
	}
	if i, e := os.Stat(pthTree); e == nil {
		t := "My Media"
		if ir {
			t = "Моя медиатека"
		}
		n := "/" + i.Name() + "/"
		pl.add(n, t, "/fy.svg", n)
	} else if !os.IsExist(e) {
		pl.Note = "<b style='color:salmon'>" + e.Error() + "</b>"
	}
	d, i := "<span style='color:salmon'>Torrserve ", "/ort.svg"
	if sets.Torrserve != "" {
		d, i = "<span style='color:lime'>"+sets.Torrserve, "/ogt.svg"
	} else if ir {
		d += "не задан!"
	} else {
		d += "is not set!"
	}
	pl.add("/torrserve?IP=TORRSERVE_IP", "Torrserve", i, d+"</span>")
	if fs, e := ioutil.ReadDir(pthPlugs); e == nil {
		for _, f := range fs {
			if n := f.Name(); f.IsDir() {
				if p, e := plugInfo(n); e == nil {
					if p.Title == "" {
						p.Title = n
					}
					n = "/" + n + "/"
					if p.Icon != "" && p.Icon[0] != '/' {
						p.Icon = n + p.Icon
					}
					pl.add(n, p.Title, p.Icon, n)
				}
			}
		}
	} else if !os.IsNotExist(e) {
		pl.Note += "; <b style='color:salmon'>" + e.Error() + "</b>"
	}
	pl.write(w)
}
func files(w http.ResponseWriter, r *http.Request) {
	p, u := filepath.Clean(r.URL.Path[1:]), r.URL.EscapedPath()
	if i, e := os.Stat(p); os.IsNotExist(e) {
		panic(404)
	} else if e != nil {
		panic(e)
	} else if !i.IsDir() {
		http.ServeFile(w, r, p)
	} else if ff, e := ioutil.ReadDir(p); e != nil {
		panic(e)
	} else {
		pl := &playlist{r: r, Cache: "nocache"}
		for _, f := range ff {
			n, m := f.Name(), f.Mode()
			if m&fs.ModeSymlink != 0 {
				if f, e = os.Stat(filepath.Join(p, n)); e == nil {
					m = f.Mode()
				} else {
					continue
				}
			}
			t, l := fileType(n), u+url.PathEscape(n)
			switch {
			case m.IsDir():
				pl.add(l+"/", n, "/fy.svg", nil)
			case !m.IsRegular():
				continue
			case t == 'l':
				pl.add(l, n, "/fyl.svg", int64(0))
			case t == 't':
				if sets.Torrserve != "" {
					pl.add("/torrserve?link="+url.QueryEscape(l), n, "/fyt.svg", int64(0))
				}
			default:
				pl.add(l, n, string(t), f.Size())
			}
		}
		pl.write(w)
	}
}
func fileType(fn string) byte {
	x := strings.ToLower(strings.TrimPrefix(filepath.Ext(fn), "."))
	for t, es := range fileTypes {
		for _, e := range es {
			if e == x {
				return t
			}
		}
	}
	return 0
}
