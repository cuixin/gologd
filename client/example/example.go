package main

import (
	"flag"
	logc "github.com/cuixin/gologd/client"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	flag.Parse()
	logc.Start("localhost:1234")
	for i := 1; i < 10000; i++ {
		time.Sleep(time.Millisecond)
		logc.Log("t_user_log", logc.P{"uid": i, "name": "jack", "operation": 1})
	}
	defer logc.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)

	for sig := range c {
		switch sig {
		case syscall.SIGINT:
			return
		}
	}
}
