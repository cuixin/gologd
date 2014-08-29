package main

import (
	"errors"
	"log"
	"net"
)

type TcpServer struct {
	addr string
	ln   *net.TCPListener
}

func (this *TcpServer) accept() (conn *net.TCPConn, err error) {
	conn, err = this.ln.AcceptTCP()
	return
}

func (this *TcpServer) buildListener() error {
	if this.ln != nil {
		return errors.New("server has started")
	}

	laddr, err := net.ResolveTCPAddr("tcp", this.addr)
	if err != nil {
		log.Fatalf("resolve local addr failed:%s\n", err.Error())
		return err
	}

	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatalf("build listener failed:%s\n", err.Error())
		return err
	}

	log.Printf("listen %s\n", this.addr)
	this.ln = ln
	return nil
}

func (this *TcpServer) closeListener() {
	if this.ln != nil {
		this.ln.Close()
	}
}
