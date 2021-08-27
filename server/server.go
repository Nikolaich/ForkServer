package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	Executable    string
	Addr          = ":8027"
	Update        = time.Hour * 24
	Restart       = func() { mutex.Lock(); os.Exit(0) }
	Err           = log.New(os.Stderr, "<!> ", log.LstdFlags)
	Wrn           = log.New(os.Stderr, "(i) ", log.LstdFlags)
	Inf           = log.New(os.Stdout, "(i) ", log.LstdFlags)
	SkipVerifyTLS bool
	server        = &http.Server{Handler: http.HandlerFunc(handler)}
	mutex         = new(sync.Mutex)
)

func Init() error { return sets.get() }
func Run() (e error) {
	var nv bool
	if e = os.Remove(Executable + ".old"); e == nil {
		Wrn.Println(Name, "has been updated to v.", Vers)
		nv = true
	} else if !os.IsNotExist(e) {
		Wrn.Println(e)
		nv = true
	}
	checkUpdate(nv, Update)
	Wrn.Println("Start", Name, "v.", Vers, "with plugins v.", sets.PlugsTag, "listen to", Addr)
	server.Addr = Addr
	if e = server.ListenAndServe(); e == http.ErrServerClosed {
		Wrn.Println(e)
		e = nil
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
