package tengosrv

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/d5/tengo/v2"
)

type server struct {
	w http.ResponseWriter
	r *http.Request
	c *http.Client
	p string
	h bool
}

func Server(w http.ResponseWriter, r *http.Request, c *http.Client, p string, vars map[string]tengo.Object) map[string]tengo.Object {
	if c == nil {
		c = &http.Client{}
	}
	s := &server{w, r, c, p, false}
	ret := map[string]tengo.Object{
		"proto":       &tengo.String{Value: r.Proto},
		"method":      &tengo.String{Value: r.Method},
		"host":        &tengo.String{Value: r.Host},
		"remote_addr": &tengo.String{Value: r.RemoteAddr},
		"header":      vals2map(r.Header),
		"uri":         &tengo.String{Value: r.RequestURI},
		"read":        &tengo.UserFunction{Name: "read", Value: s.read},
		"write":       &tengo.UserFunction{Name: "write", Value: s.write},
		"request":     &tengo.UserFunction{Name: "request", Value: s.request},
		"file":        &tengo.UserFunction{Name: "file", Value: s.file},
	}
	for k, v := range vars {
		ret[k] = v
	}
	return ret
}
func (s *server) read(args ...tengo.Object) (r tengo.Object, e error) {
	switch len(args) {
	case 0:
		if bs, er := ioutil.ReadAll(s.r.Body); er == nil {
			r = &tengo.Bytes{Value: bs}
		} else {
			r = &tengo.Error{Value: &tengo.String{Value: er.Error()}}
		}
	case 1:
		switch a := args[0].(type) {
		case *tengo.String:
			r = &tengo.String{Value: s.r.FormValue(a.Value)}
		case *tengo.Bool:
			if er := s.r.ParseForm(); er != nil {
				r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
			} else if a.IsFalsy() {
				r = vals2map(s.r.PostForm)
			} else {
				r = vals2map(s.r.Form)
			}
		default:
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/bool", Found: args[0].TypeName()}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}
func (s *server) write(args ...tengo.Object) (tengo.Object, error) {
	c := 0
	for n, arg := range args {
		switch a := arg.(type) {
		case *tengo.Map:
			if s.h {
				return nil, tengo.ErrInvalidArgumentType{Name: strconv.Itoa(n) + "-th", Expected: "nor map or int", Found: a.TypeName()}
			}
			for k, vs := range map2vals(a) {
				for _, v := range vs {
					s.w.Header().Add(k, v)
				}
			}
		case *tengo.Int:
			if s.h {
				return nil, tengo.ErrInvalidArgumentType{Name: strconv.Itoa(n) + "-th", Expected: "nor map or int", Found: a.TypeName()}
			}
			s.h, c = true, int(a.Value)
		default:
			var e error
			if c > 0 {
				if v, o := tengo.ToString(a); !o {
					s.w.WriteHeader(c)
				} else if c < 300 {
					s.w.WriteHeader(c)
					_, e = s.w.Write([]byte(v))
				} else if c < 400 {
					http.Redirect(s.w, s.r, v, c)
				} else {
					http.Error(s.w, v, c)
				}
			} else if v, o := a.(*tengo.Bytes); o {
				_, e = s.w.Write(v.Value)
			} else if v, o := tengo.ToString(a); o {
				_, e = s.w.Write([]byte(v))
			}
			if e != nil {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			} else {
				c, s.h = 0, true
			}
		}
	}
	return nil, nil
}
func (s *server) request(args ...tengo.Object) (tengo.Object, error) {
	var (
		ctp string
		met = "GET"
		brd io.Reader
		opt *tengo.Map
		req *http.Request
	)
	switch len(args) {
	case 2:
		switch v := args[1].(type) {
		case *tengo.String:
			met = v.Value
		case *tengo.Bytes:
			brd, met, ctp = bytes.NewReader(v.Value), "POST", "text/plain"
		case *tengo.Map:
			opt = v
			if o, k := opt.Value["body"]; k {
				met, ctp = "POST", "text/plain"
				switch v := o.(type) {
				case *tengo.Map:
					brd, ctp = strings.NewReader(url.Values(map2vals(v)).Encode()), "application/x-www-form-urlencoded"
				case *tengo.Bytes:
					brd = bytes.NewReader(v.Value)
				default:
					s, _ := tengo.ToString(o)
					brd = strings.NewReader(s)
				}
			}
			if o, k := opt.Value["method"]; k {
				met, _ = tengo.ToString(o)
				met = strings.ToUpper(met)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{Name: "second", Expected: "map/bytes/string", Found: args[1].TypeName()}
		}
		fallthrough
	case 1:
		var e error
		s, _ := tengo.ToString(args[0])
		if req, e = http.NewRequest(met, s, brd); e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		} else if ctp != "" {
			req.Header.Set("Content-Type", ctp)
		}
	default:
		return nil, tengo.ErrWrongNumArguments
	}
	if opt != nil {
		if o, k := opt.Value["query"]; k {
			if q, k := o.(*tengo.Map); k {
				req.URL.RawQuery = url.Values(map2vals(q)).Encode()
			} else if s, _ := tengo.ToString(o); s != "" {
				req.URL.RawQuery = s
			}
		}
		if o, k := opt.Value["header"].(*tengo.Map); k {
			for h, hv := range map2vals(o) {
				req.Header[h] = hv
			}
		}
		if o, k := opt.Value["cookies"].(*tengo.Array); k {
			for _, v := range o.Value {
				if m, k := v.(*tengo.Map); k {
					if n, _ := tengo.ToString(m.Value["name"]); n != "" {
						s, _ := tengo.ToString(m.Value["value"])
						req.AddCookie(&http.Cookie{Name: n, Value: s})
					}
				}
			}
		}
		if u, k := opt.Value["user"].(*tengo.String); k && u.Value != "" {
			if p, k := opt.Value["pass"].(*tengo.String); k && p.Value != "" {
				req.SetBasicAuth(u.Value, p.Value)
			}
		}
		if o, k := opt.Value["follow"]; k {
			if o.IsFalsy() {
				s.c.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			}
		}
		if o, k := opt.Value["timeout"].(*tengo.Int); k && o.Value > 0 {
			s.c.Timeout = time.Duration(o.Value) * time.Second
		}
	}
	rsp, e := s.c.Do(req)
	if e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	}
	defer rsp.Body.Close()
	bs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	}
	cs := new(tengo.Array)
	for _, cc := range rsp.Cookies() {
		cs.Value = append(cs.Value, &tengo.Map{Value: map[string]tengo.Object{"name": &tengo.String{Value: cc.Name}, "value": &tengo.String{Value: cc.Value}}})
	}
	au, ap, _ := rsp.Request.BasicAuth()
	return &tengo.Map{Value: map[string]tengo.Object{
		"status":  &tengo.Int{Value: int64(rsp.StatusCode)},
		"user":    &tengo.String{Value: au},
		"pass":    &tengo.String{Value: ap},
		"header":  vals2map(rsp.Header),
		"cookies": cs,
		"body":    &tengo.Bytes{Value: bs},
		"size":    &tengo.Int{Value: rsp.ContentLength},
		"url":     &tengo.String{Value: rsp.Request.URL.String()},
	}}, nil
}
func (s *server) file(args ...tengo.Object) (tengo.Object, error) {
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
		p = filepath.Join(s.p, filepath.Clean(n.Value))
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
