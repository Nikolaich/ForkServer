package server

import (
	"crypto/tls"
	"net/http"
	"os"
	"sync"
	"time"
)

const Name, Vers = "ForkServer", "0.05"

var (
	Executable           string
	Addr                 = ":8027"
	Update               = time.Hour * 24
	Restart              = func() { Wait(); os.Exit(0) }
	Error, Warning, Info func(...interface{})
	SkipVerifyTLS        bool
	mutex                = new(sync.Mutex)
)

func Init() error { return sets.get() }
func Run() error {
	var nv bool
	if e := os.Remove(Executable + ".old"); e == nil {
		Warning(Name, "has been updated to v.", Vers)
		nv = true
	} else if !os.IsNotExist(e) {
		Warning(e)
		nv = true
	}
	checkUpdate(nv, Update)
	Warning("Start", Name, "v.", Vers, "with plugins v.", sets.PlugsTag, "listen to", Addr)
	return http.ListenAndServe(Addr, http.HandlerFunc(handler))
}
func Wait() { mutex.Lock() }
func httpClient() *http.Client {
	if SkipVerifyTLS {
		return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: SkipVerifyTLS}}}
	} else {
		return &http.Client{}
	}
}
