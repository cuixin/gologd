package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
)

type Params map[string]interface{}

type TransData struct {
	Table  string
	Params Params
}

const (
	HEADER_SIZE = 4
	PAYLOAD_MAX = 64 * 1024 // 64K
)

type Server struct {
	TcpServer
	wg sync.WaitGroup
}

func (this *Server) listen() {
	defer this.wg.Done()
	for {
		conn, err := this.accept()
		if err != nil {
			log.Printf("comet accept failed:%s\n", err.Error())
			break
		}
		this.wg.Add(1)
		go this.handleClient(conn)
	}
}

func (this *Server) Start() error {
	err := this.buildListener()
	if err != nil {
		return err
	}

	this.wg.Add(1)
	go this.listen()
	return nil
}

func (this *Server) handleClient(conn *net.TCPConn) {
	defer this.wg.Done()
	defer conn.Close()
	addr, err := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		log.Printf("ResolveTCPAddr failed [%v], %v\n", conn.RemoteAddr(), err)
		return
	}
	log.Printf("New client [%s]\n", addr.IP.String())
	header := make([]byte, HEADER_SIZE)
	buf := make([]byte, PAYLOAD_MAX)
	for {
		// read header : 4-bytes
		n, err := io.ReadFull(conn, header)
		if n == 0 && err == io.EOF {
			break
		} else if err != nil {
			log.Printf("[%s] error receiving header:%s\n", conn.RemoteAddr().String(), err)
			break
		}

		// read payload, the size of the payload is given by header
		size := binary.LittleEndian.Uint32(header)
		if size > PAYLOAD_MAX {
			// 数据意外的过长，扔掉本次消息
			log.Printf("[%v] messge too long [%d] bytes, so discard it.\n",
				conn.RemoteAddr(), size)
			_, err = io.CopyN(ioutil.Discard, conn, int64(size))
			if err != nil {
				break
			}
			continue
		}

		data := buf[:size]
		n, err = io.ReadFull(conn, data)

		if err != nil {
			log.Printf("error receiving payload:%s\n", err)
			break
		}
		MySQLHandler(data)
		conn.Write([]byte{1})
	}
	log.Printf("Client [%s] closed\n", addr.IP.String())
}

func (self *Server) Stop() {
	self.closeListener()
}

func NewServer(addr string) *Server {
	server := new(Server)
	server.TcpServer.addr = addr
	return server
}

func Insert(tab string, params Params) error {
	fields_str := ""
	values_str := ""
	values := make([]interface{}, 0)
	for k, v := range params {
		fields_str += k + ","
		values_str += "?,"
		values = append(values, v)
	}
	fields_str = strings.TrimRight(fields_str, ",")
	values_str = strings.TrimRight(values_str, ",")
	build_sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", tab, fields_str, values_str)
	// log.Println(build_sql)
	stmt, err := db.Prepare(build_sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, e := stmt.Exec(values...); e != nil {
		return e
	}
	return nil
}

func MySQLHandler(data []byte) {
	T := new(TransData)
	json.Unmarshal(data, T)
	if e := Insert(T.Table, T.Params); e != nil {
		log.Println(e)
	}
}
