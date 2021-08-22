//go:build windows
//+build windows

package main

import (
	"ForkServer/server"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/judwhite/go-svc"
)

type service struct{ IsWindowsService bool }

func (s *service) Init(env svc.Environment) (e error) {
	if s.IsWindowsService = env.IsWindowsService(); s.IsWindowsService {
		server.Restart = s.restart
		if fileErr, e = os.Create("errors.log"); e == nil {
			if fileInf != nil {
				fileInf, e = os.Create("info.log")
			}
		}
	}
	return
}
func (s *service) Start() (e error) {
	if e = server.Init(); e == nil {
		go s.run()
	}
	return
}
func (s *service) Stop() error {
	defer s.close()
	server.Wait()
	return nil
}
func (s *service) run() {
	defer s.close()
	if e := server.Run(); e != nil {
		server.Error(e)
		os.Exit(1)
	}
}
func (s *service) close() {
	if s.IsWindowsService {
		fileErr.Close()
		if fileInf != nil {
			fileInf.Close()
		}
	}
}
func (s *service) restart() {
	if e := exec.Command("cmd", "/C", "net stop "+server.Name+" && net start "+server.Name).Start(); e != nil {
		server.Error(e)
	}
}
func runService(cmd string) {
	switch cmd {
	case "install":
		d := filepath.Dir(server.Executable)
		check(exec.Command("sc", "create", server.Name, "binpath=", server.Executable+" -d "+d, "start=", "auto", "DisplayName=", server.Name).Run())
		check(exec.Command("sc", "description", server.Name, server.Name+" for ForkPlayer").Run())
		check(exec.Command("net", "start", server.Name).Run())
		check(server.Name + " service installed and started!")
		fallthrough
	case "treeview":
		check("Create the symbolic link \"treeview\" to your media folder!")
	case "uninstall":
		check(exec.Command("net", "stop", server.Name).Run())
		check(exec.Command("sc", "delete", server.Name).Run())
		check(server.Name + " service stopped and deleted!")
	case "/?", "?", "-h", "-help", "--help":
		check("Usage: " + os.Args[0] + " [cmd]\nwhere [cmd] can be one of: install, uninstall, treeview")
	default:
		if e := svc.Run(new(service)); e != nil {
			server.Error(e)
			os.Exit(1)
		}
	}
}
