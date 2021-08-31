package main

import (
	"ForkServer/server"
	"fmt"
	"log"
	"os/exec"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

type service struct {
	Name, WD, Addr string
	Log            *eventlog.Log
	Err            chan error
}

func init() {
	runService = func(sn, la, wd string, li bool) {
		if is, e := svc.IsWindowsService(); e != nil {
			log.Fatalln(e)
		} else if !is {
			runSTD(sn, la, wd, li)
		} else if lg, e := eventlog.Open(server.Name); e != nil {
			log.Fatalln(e)
		} else {
			s := &service{sn, wd, la, lg, make(chan error)}
			if li {
				server.Info = s.inf
			}
			svc.Run(s.Name, s)
		}
	}
}

func (s *service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	server.Error = s.err
	server.Warning = s.wrn
	server.Restart = s.restart
	go func() { s.Err <- server.Run(s.Addr, s.WD) }()
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case e := <-s.Err:
			if e != nil {
				s.err(e)
				return true, 1
			}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				if e := server.Stop(); e != nil {
					s.err(e)
					return true, 2
				}
				return false, 0
			default:
				s.Log.Error(1, fmt.Sprintf("unexpected windows control request #%d", c))
			}
		}
	}
}
func (s *service) err(v ...interface{}) { s.Log.Error(1, fmt.Sprint(v...)) }
func (s *service) wrn(v ...interface{}) { s.Log.Warning(1, fmt.Sprint(v...)) }
func (s *service) inf(v ...interface{}) { s.Log.Info(1, fmt.Sprint(v...)) }
func (s *service) restart() {
	if e := exec.Command("cmd", "/C", "net stop "+s.Name+" && net start "+s.Name).Start(); e != nil {
		s.err(e)
	}
}
