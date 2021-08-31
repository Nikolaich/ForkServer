package server

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	Update        = time.Hour * 24
	Restart       = func() { mutex.Lock(); os.Exit(0) }
	Error         = func(v ...interface{}) {}
	Warning       = func(v ...interface{}) {}
	Info          = func(v ...interface{}) {}
	SkipVerifyTLS bool

	server = &http.Server{Addr: ":8027", Handler: http.HandlerFunc(handler)}
	mutex  = new(sync.Mutex)
	exec   string
)

func Run(addr, wd string) (e error) {
	var nv bool
	if addr != "" {
		server.Addr = addr
	}
	if exec, e = os.Executable(); e == nil {
		if wd == "" {
			wd = filepath.Dir(exec)
		}
		if e = os.Chdir(wd); e == nil {
			if e = os.Remove(exec + ".old"); e == nil {
				Warning(Name, "has been updated to v.", Vers)
				nv = true
			} else if !os.IsNotExist(e) {
				Error(e)
				nv = true
			}
			if e = sets.get(); e == nil {
				checkUpdate(nv, Update)
				Warning("Start", Name, "v.", Vers, "with plugins v.", sets.PlugsTag, "listen to", server.Addr)
				if e = server.ListenAndServe(); e == http.ErrServerClosed {
					Warning(e)
					e = nil
				}
			}
		}
	}
	return
}
func Stop() error {
	mutex.Lock()
	return server.Close()
}
func httpClient() *http.Client {
	if SkipVerifyTLS {
		return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: SkipVerifyTLS}}}
	} else {
		return &http.Client{}
	}
}
