package main

import (
	"ForkServer/server"
	"os"
	"os/exec"

	"github.com/judwhite/go-svc"
)

type service struct {
	Name  string
	Debug bool
	Files []*os.File
}

func init() {
	serve = func(svn string, dbg bool) {
		s := &service{svn, dbg, nil}
		if e := svc.Run(s); e != nil {
			s.close(e)
		}
	}
}
func (s *service) Init(env svc.Environment) (e error) {
	if env.IsWindowsService() {
		var f *os.File
		server.Restart = s.restart
		if f, e = os.Create("errors.log"); e == nil {
			s.Files = append(s.Files, f)
			server.Err.SetOutput(s.Files[0])
			server.Wrn.SetOutput(s.Files[0])
			if s.Debug {
				if f, e = os.Create("info.log"); e == nil {
					s.Files = append(s.Files, f)
					server.Inf.SetOutput(s.Files[1])
				}
			}
		}
	}
	return
}
func (s *service) Start() (e error) {
	if e = server.Init(); e == nil {
		go s.close(server.Run())
	}
	return
}
func (s *service) Stop() error {
	return server.Stop()
}
func (s *service) close(e error) {
	for _, f := range s.Files {
		f.Close()
	}
	if e != nil {
		server.Err.Fatalln(e)
	}
}
func (s *service) restart() {
	if e := exec.Command("cmd.exe", "/C", "net stop "+s.Name+" && net start "+s.Name).Start(); e != nil {
		server.Err.Println(e)
	}
}
