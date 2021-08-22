package main

import (
	"ForkServer/server"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

var fileErr, fileInf = os.Stderr, os.Stdout

func check(m interface{}) {
	switch v := m.(type) {
	case string:
		os.Stdout.WriteString(v + "\n")
	case error:
		os.Stderr.WriteString(v.Error() + "\n")
		os.Exit(1)
	case int:
		os.Exit(v)
	}
}
func main() {
	var e error
	server.Executable, e = os.Executable()
	check(e)
	f, c := log.LstdFlags, ""
	if l := len(os.Args); l > 1 {
		o := ""
		for _, a := range os.Args[1:] {
			if o == "" {
				switch a {
				case "-a", "-d", "-u":
					o = a
				case "-s":
					server.SkipVerifyTLS = true
				case "-t":
					f = 0
				case "-i":
					fileInf = nil
				default:
					c = a
				}
			} else {
				switch o {
				case "-a":
					server.Addr = a
				case "-d":
					check(os.Chdir(a))
				case "-u":
					u, _ := strconv.Atoi(a)
					server.Update = time.Duration(u) * time.Hour
				}
				o = ""
			}
		}
	}
	server.Error = log.New(fileErr, "<!> ", f).Println
	server.Warning = log.New(fileErr, "(i) ", f).Println
	if fileInf == nil {
		server.Info, fileInf = log.New(ioutil.Discard, "(i) ", f).Println, nil
	} else {
		server.Info = log.New(fileInf, "(i) ", f).Println
	}
	runService(c)
}
