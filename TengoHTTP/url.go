package tengohttp

import (
	"net/url"

	tengo "github.com/d5/tengo/v2"
)

var URL = map[string]tengo.Object{
	"encode": &tengo.UserFunction{Name: "encode", Value: encode},
	"decode": &tengo.UserFunction{Name: "decode", Value: decode},
	"url":    &tengo.UserFunction{Name: "url", Value: parseurl},
	"query":  &tengo.UserFunction{Name: "query", Value: query},
}

func parseurl(args ...tengo.Object) (r tengo.Object, e error) {
	if len(args) != 1 {
		e = tengo.ErrWrongNumArguments
	} else {
		switch a := args[0].(type) {
		case *tengo.String:
			if u, e := url.Parse(a.Value); e != nil {
				r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
			} else {
				ret := &tengo.Map{Value: map[string]tengo.Object{
					"scheme":       &tengo.String{Value: u.Scheme},
					"opaque":       &tengo.String{Value: u.Opaque},
					"user":         &tengo.String{Value: u.User.Username()},
					"host":         &tengo.String{Value: u.Host},
					"path":         &tengo.String{Value: u.Path},
					"raw_path":     &tengo.String{Value: u.EscapedPath()},
					"query":        &tengo.String{Value: u.RawQuery},
					"fragment":     &tengo.String{Value: u.Fragment},
					"raw_fragment": &tengo.String{Value: u.EscapedFragment()},
				}}
				if p, o := u.User.Password(); o {
					ret.Value["pass"] = &tengo.String{Value: p}
				}
				r = ret
			}
		case *tengo.Map:
			u, au, ap := new(url.URL), "", ""
			for k, v := range a.Value {
				if s, _ := tengo.ToString(v); s != "" {
					switch k {
					case "scheme":
						u.Scheme = s
					case "opaque":
						u.Opaque = s
					case "user":
						au = s
					case "pass":
						ap = s
					case "host":
						u.Host = s
					case "path":
						u.Path = s
					case "query":
						u.RawQuery = s
					case "fragment":
						u.Fragment = s
					}
				}
			}
			if au != "" {
				if ap != "" {
					u.User = url.UserPassword(au, ap)
				} else {
					u.User = url.User(au)
				}
			}
			r = &tengo.String{Value: u.String()}
		default:
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
		}
	}
	return
}
func encode(args ...tengo.Object) (r tengo.Object, e error) {
	pth := false
	switch len(args) {
	case 2:
		pth = !args[1].IsFalsy()
		fallthrough
	case 1:
		if s, o := args[0].(*tengo.String); !o {
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
		} else if pth {
			r = &tengo.String{Value: url.PathEscape(s.Value)}
		} else {
			r = &tengo.String{Value: url.QueryEscape(s.Value)}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}
func decode(args ...tengo.Object) (r tengo.Object, e error) {
	pth := false
	switch len(args) {
	case 2:
		pth = !args[1].IsFalsy()
		fallthrough
	case 1:
		if s, o := args[0].(*tengo.String); !o {
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
		} else {
			var rs string
			if pth {
				rs, e = url.PathUnescape(s.Value)
			} else {
				rs, e = url.QueryUnescape(s.Value)
			}
			if e != nil {
				r, e = &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			} else {
				r = &tengo.String{Value: rs}
			}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}
func query(args ...tengo.Object) (r tengo.Object, e error) {
	if len(args) != 1 {
		e = tengo.ErrWrongNumArguments
	} else if a, o := args[0].(*tengo.String); o {
		if vs, er := url.ParseQuery(a.Value); er != nil {
			r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
		} else {
			r = vals2map(vs)
		}
	} else if a, o := args[0].(*tengo.Map); o {
		u := new(url.Values)
		for k, v := range a.Value {
			if ao, o := v.(*tengo.Array); o {
				for _, av := range ao.Value {
					s, _ := tengo.ToString(av)
					u.Add(k, s)
				}
			} else {
				s, _ := tengo.ToString(v)
				u.Set(k, s)
			}
		}
		r = &tengo.String{Value: u.Encode()}
	} else {
		e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
	}
	return
}
func vals2map(vals map[string][]string) *tengo.Map {
	r := &tengo.Map{Value: make(map[string]tengo.Object)}
	for k, vs := range vals {
		a := &tengo.Array{}
		for _, v := range vs {
			a.Value = append(a.Value, &tengo.String{Value: v})
		}
		r.Value[k] = a
	}
	return r
}
