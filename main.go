package main

import (
	"ForkServer/server"
	"log"
	"os"
	"strconv"
	"time"
)

var runService = runSTD

func runSTD(sn, la, wd string, li bool) {
	log.SetPrefix("Err: ")
	server.Error = log.Println
	server.Warning = log.New(os.Stderr, "Wrn: ", log.Flags()).Println
	if li {
		server.Info = log.New(os.Stdout, "Inf: ", log.Flags()).Println
	}
	if e := server.Run(la, wd); e != nil {
		log.Fatalln(e)
	}
}
func main() {
	li, wd, la, sn := true, "", "", server.Name
	if len(os.Args) > 1 {
		o := ""
		for _, a := range os.Args {
			switch o {
			case "-a":
				la, o = a, ""
			case "-d":
				wd, o = a, ""
			case "-n":
				sn, o = a, ""
			case "-u":
				o = ""
				if h, e := strconv.Atoi(a); e == nil {
					server.Update = time.Hour * time.Duration(h)
				}
			default:
				switch a {
				case "-a", "-d", "-u", "-n":
					o = a
				case "-s":
					server.SkipVerifyTLS = true
				case "-i":
					li = false
				}
			}
		}
	}
	runService(sn, la, wd, li)
}
