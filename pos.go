//go:build !windows
//+build !windows

package main

import (
	"ForkServer/server"
	"os"
)

func runService(cmd string) {
	e := server.Init()
	if e == nil {
		e = server.Run()
	}
	if e != nil {
		server.Error(e)
		os.Exit(1)
	}
}
