package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type torrfile struct {
	ID     int
	Path   string
	Length int64
}
type torrinfo struct {
	Name, Hash string
	Stat       int
	PeersT     int   `json:"total_peers"`
	PeersA     int   `json:"active_peers"`
	Seeders    int   `json:"connected_seeders"`
	Length     int64 `json:"torrent_size"`
	//Files      []torrfile `json:"file_stats"`
}

func checkTorr(a string) (v string, e error) {
	if a != "" {
		var r *http.Response
		if r, e = http.Get("http://" + a + "/echo"); e == nil {
			var b []byte
			if b, e = ioutil.ReadAll(r.Body); e == nil {
				if r.StatusCode == 200 {
					if v = string(b); !strings.HasPrefix(v, "MatriX") {
						e = errors.New("Version " + v + " is not supported!")
					}
				} else {
					e = errors.New(a + "/echo answered: " + r.Status)
				}
			}
			r.Body.Close()
		}
	}
	return
}
func torrCurr(pl *playlist) {
	var ts []torrinfo
	r, e := http.Post("http://"+sets.Torrserve+"/torrents", "application/json", strings.NewReader(`{"action":"list"}`))
	check(e)
	defer r.Body.Close()
	if r.StatusCode != 200 {
		ioutil.ReadAll(r.Body)
		panic(r.StatusCode)
	}
	check(json.NewDecoder(r.Body).Decode(&ts))
	pl.Cache = "nocache"
	for _, t := range ts {
		i, d := "/fyt.svg", "<b>"+t.Name+"</b><p>"+formatBytes(t.Length)
		if t.Stat > 2 {
			i = "/fgt.svg"
			d += "</p><p style='color:lime'><img style='height:.8em' src='http://" + pl.r.Host + "/d2.svg'> " + strconv.Itoa(t.PeersA) + " / " + strconv.Itoa(t.PeersT)
			if t.Seeders > 0 {
				d += "<br><img style='height:.8em' src='http://" + pl.r.Host + "/d.svg'> " + strconv.Itoa(t.Seeders)
			}
		}
		pl.add("/torrserve?link="+t.Hash, t.Name, i, d+"</p><div style='font-size:small;color:gray'>"+t.Hash+"</div>")
	}
}
func torrLink(pl *playlist, ul string) {
	var t struct {
		torrinfo
		Files []torrfile `json:"file_stats"`
	}
	u := "http://" + sets.Torrserve
	if ul[0] == '/' {
		ul = "http://" + pl.r.Host + ul
	}
	ul = url.QueryEscape(ul)
	r, e := http.Get(u + "/stream/?stat&link=" + ul)
	check(e)
	defer r.Body.Close()
	if r.StatusCode != 200 {
		ioutil.ReadAll(r.Body)
		panic(r.StatusCode)
	}
	check(json.NewDecoder(r.Body).Decode(&t))
	//if t.Stat > 2 {
	//pl.Name += "<b style='color:limegreen'> <img style='height:.8em' src='http://" + pl.r.Host + "/d2g.svg'> " + strconv.Itoa(t.PeersA) + " / " + strconv.Itoa(t.PeersT) + "</b>"
	//}
	cp := ""
	for _, f := range t.Files {
		var i int
		if i = strings.IndexByte(f.Path, '/'); i > 0 {
			f.Path = f.Path[i+1:]
			if i = strings.LastIndexByte(f.Path, '/'); i > 0 {
				if cp == f.Path[:i] {
					f.Path = f.Path[i+1:]
				} else {
					cp = f.Path[:i]
				}
			}
		}
		pl.add(u+"/stream/video?play&link="+ul+"&index="+strconv.Itoa(f.ID), f.Path, string(fileType(f.Path[i+1:])), f.Length)
	}
}
func torrserve(w http.ResponseWriter, r *http.Request) {
	pl := &playlist{r: r}
	if a := r.FormValue("IP"); a != "" && sets.Torrserve == "" {
		a += ":8090"
		if v, e := checkTorr(a); e == nil {
			sets.Torrserve, pl.Note = a, v
			check(sets.put())
		}
	}
	if sets.Torrserve == "" {
		t := "Torrserve "
		if r.FormValue("w_lang") == "ru" {
			t += "не найден!"
		} else {
			t += "is not found!"
		}
		pl.add("cmd:historyback(1)", t, "/d1r.svg", nil)
		pl.add("/set?Torrserve=search", "Torrserve IP[:PORT]", "/ort.svg", true)
	} else if l := r.FormValue("link"); l != "" {
		torrLink(pl, l)
	} else {
		torrCurr(pl)
		pl.add("/set?Torrserve=search", "Torrserve "+sets.Torrserve, "/ogt.svg", true)
	}
	pl.write(w)
}
