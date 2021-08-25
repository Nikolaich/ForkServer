package server

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type plistItem map[string]string
type playlist struct {
	r     *http.Request
	Cache string      `json:"cacheinfo,omitempty"`
	Type  string      `json:"typeList,omitempty"`
	Note  string      `json:"notify,omitempty"`
	Info  string      `json:"info,omitempty"`
	Cmd   string      `json:"cmd,omitempty"`
	Menu  []plistItem `json:"menu,omitempty"`
	Pls   []plistItem `json:"channels,omitempty"`
	Pls2  []plistItem `json:"-"`
	Str   []plistItem `json:"-"`
}

func (p *playlist) add(url, ttl, icn string, inf interface{}) {
	i, l, m, p2 := plistItem{"title": ttl}, "playlist_url", false, false
	switch v := inf.(type) {
	case int64:
		s, ns := strings.LastIndexByte(ttl, '/'), p.r.FormValue("shuffle") == ""
		i["title"] = ttl[s+1:]
		d := "<p>" + ttl[s+1:]
		if v > 0 {
			ir := p.r.FormValue("w_lang") == "ru"
			d += "<hr>" + formatBytes(v)
			l = "stream_url"
			switch icn {
			case "a":
				icn = "/ag.svg"
				if ir {
					i["group"] = "Аудио"
				} else {
					i["group"] = "Audio"
				}
			case "v":
				icn = "/vg.svg"
				if ir {
					i["group"] = "Видео"
				} else {
					i["group"] = "Video"
				}
			default:
				icn = ""
			}
		}
		if s > 0 && ns {
			i["before"] = "<div style='color:yellow;text-align:left'><img src='http://" + p.r.Host + "/fy.svg' style='height:1em'>&nbsp;" + ttl[:s] + "</div>"
		}
		i["description"], p2 = d+"</p>", true
	case string:
		i["description"] = v
	case bool:
		m = true
		if v {
			i["search_on"] = "search"
		}
	}
	if icn != "" {
		if icn[0] == '/' {
			i["logo_30x30"] = "http://" + p.r.Host + icn
		} else {
			i["logo_30x30"] = icn
		}
	}
	if url[0] == '/' {
		i[l] = "http://" + p.r.Host + url
	} else {
		i[l] = url
	}
	if m {
		p.Menu = append(p.Menu, i)
	} else if l == "stream_url" {
		p.Str = append(p.Str, i)
	} else if p2 {
		p.Pls2 = append(p.Pls2, i)
	} else {
		p.Pls = append(p.Pls, i)
	}
}
func (p *playlist) write(w io.Writer) error {
	p.Pls, p.Pls2 = append(p.Pls, p.Pls2...), nil
	if l := len(p.Str); l > 0 {
		if l > 1 {
			t := "Shuffle"
			if p.r.FormValue("w_lang") == "ru" {
				t = "Перетасовать"
			}
			u := p.r.FormValue("link")
			if u != "" {
				u = "&link=" + url.QueryEscape(u)
			}
			p.add(p.r.URL.EscapedPath()+"?shuffle=true"+u, "<span style='color:lime'>"+t+"</b>", "/sy.svg", false)
			if p.r.FormValue("shuffle") != "" {
				rand.Seed(time.Now().UnixNano())
				rand.Shuffle(l, func(i, j int) { p.Str[i], p.Str[j] = p.Str[j], p.Str[i] })
			}
		}
		p.Pls = append(p.Pls, p.Str...)
	} else if len(p.Pls) == 0 && p.Cmd == "" {
		t := "There is nothing!"
		if p.r.FormValue("w_lang") == "ru" {
			t = "Здесь пусто!"
		}
		p.add("cmd:historyback(1);", t, "/d1r.svg", nil)
	}
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	return j.Encode(p)
}
func formatBytes(i int64) string {
	f, b := float64(i), ""
	for _, b = range []string{" B", " KB", " MB", " GB", " TB"} {
		if f < 1000 {
			break
		} else {
			f = f / 1024
		}
	}
	return strconv.FormatFloat(f, 'f', 2, 64) + b
}
