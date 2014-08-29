package client

import (
	"encoding/json"
	"github.com/golang/glog"
	"time"
)

type P map[string]interface{}

type TransData struct {
	Table  string
	Params P
}

var channel = make(chan *TransData, 1024)
var closeOk = make(chan struct{})

const MAX_PAYLOAD = 64 * 1024

var srv *LogServer

func Start(addr string) {
	if srv == nil {
		srv = NewLogServer(addr)
		dialErr := srv.Dail()
		if dialErr != nil {
			glog.Fatalln(dialErr.Error())
		}
		go func() {
			for v := range channel {
				data, err := json.Marshal(v)
				if err != nil {
					glog.Errorln(v.Table, v.Params)
				}
				if len(data) > MAX_PAYLOAD {
					glog.Errorln("overflow the limit", len(data))
					continue
				}
			Resend:
				sendErr := srv.Send(data)
				if sendErr != nil {
					glog.Errorln(srv.remoteAddr, "send error", sendErr.Error(), v)
				Reconnect:
					if dialErr := srv.Dail(); dialErr != nil {
						time.Sleep(1 * time.Second)
						goto Reconnect
					} else {
						glog.Errorln("Resend", v)
						goto Resend
					}
				}
				recvErr := srv.Recv()
				if recvErr != nil {
					glog.Errorln(srv.remoteAddr, "recv error", recvErr.Error(), v)
					goto Resend
				}
			}
			closeOk <- struct{}{}
		}()
	}
}

func Log(tab string, p P) {
	t := &TransData{tab, p}
	channel <- t
}

func Close() {
	close(channel)
	<-closeOk
}
