package server

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

func parserlink(w http.ResponseWriter, r *http.Request) {
	var (
		p   []string
		i   bool
		rsp *http.Response
		e   error
	)
	if r.Method == "POST" {
		b, e := ioutil.ReadAll(r.Body)
		check(e)
		p = append(p, string(b))
	} else {
		p = append(p, r.URL.RawQuery)
	}
	if s, e := url.QueryUnescape(p[0]); e == nil {
		p[0] = s
	} else {
		panic(e)
	}
	p = strings.Split(p[0], "|")
	if strings.HasPrefix(p[0], "curl") {
		var (
			l bool
			d string
			h = make(http.Header)
			q *http.Request
		)
		p[0] = p[0][strings.IndexRune(p[0], ' ')+1:]
		c, e := shlex.Split(p[0])
		check(e)
		p[0] = ""
		for j := 0; j < len(c); j++ {
			switch c[j] {
			case "-L":
				l = true
			case "-i":
				i = true
			case "-H":
				if j++; j < len(c) {
					if s := strings.SplitN(c[j], ":", 2); len(s) == 2 {
						h.Add(s[0], s[1])
					}
				}
			case "-d", "--data":
				if j++; j < len(c) {
					d = c[j]
				}
			default:
				p[0] = c[j]
			}
		}
		if d == "" {
			q, e = http.NewRequest("GET", p[0], nil)
		} else {
			q, e = http.NewRequest("POST", p[0], strings.NewReader(d))
		}
		check(e)
		for k, vs := range h {
			for _, v := range vs {
				q.Header.Add(k, v)
			}
		}
		cln := httpClient()
		if l {
			cln.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
		}
		rsp, e = cln.Do(q)
	} else {
		rsp, e = http.Get(p[0])
	}
	check(e)
	defer rsp.Body.Close()
	if i {
		for k, vs := range rsp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=\"UTF-8\"")
	}
	if l := len(p); l == 1 {
		io.Copy(w, rsp.Body)
	} else {
		b, e := ioutil.ReadAll(rsp.Body)
		check(e)
		p = append(p, "")
		if !strings.Contains(p[1], ".*?") && !strings.HasPrefix(p[2], ".*?") {
			p[1], p[2] = regexp.QuoteMeta(p[1]), regexp.QuoteMeta(p[2])
		}
		if c := regexp.MustCompile("(?s)" + p[1] + "(.*?)" + strings.TrimPrefix(p[2], ".*?")).FindSubmatch(b); len(c) > 1 {
			b = nil
			for _, cc := range c[1:] {
				b = append(b, cc...)
			}
		}
		if w.Header().Get("Content-Length") != "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		}
		w.Write(b)
	}
}
