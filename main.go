package main

import (
	"ForkServer/server"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

var serve = func(svn string, dbg bool) {
	var e error
	if e = server.Init(); e == nil {
		e = server.Run()
	}
	if e != nil {
		server.Err.Fatalln(e)
	}
}

func main() {
	var e error
	if server.Executable, e = os.Executable(); e != nil {
		server.Err.Fatalln(e)
	}
	svn, dbg := server.Name, true
	if len(os.Args) > 0 {
		o := ""
		for _, a := range os.Args[1:] {
			if o == "" {
				switch a {
				case "-a", "-d", "-u", "-n":
					o = a
				case "-s":
					server.SkipVerifyTLS = true
				case "-t":
					server.Err.SetFlags(0)
					server.Wrn.SetFlags(0)
					server.Inf.SetFlags(0)
				case "-i":
					server.Inf.SetOutput(ioutil.Discard)
					dbg = false
				}
			} else {
				switch o {
				case "-a":
					server.Addr = a
				case "-d":
					if e = os.Chdir(a); e != nil {
						server.Err.Fatalln(e)
					}
				case "-u":
					u, _ := strconv.Atoi(a)
					server.Update = time.Duration(u) * time.Hour
				case "-n":
					svn = a
				}
				o = ""
			}
		}
	}
	serve(svn, dbg)
}
