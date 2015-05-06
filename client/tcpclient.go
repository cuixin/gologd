package client

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/golang/glog"
)

const (
	HEADER_SIZE = 4
	PAYLOAD_MAX = 1024
)

type LogServer struct {
	remoteAddr string
	conn       *net.TCPConn
	recvBuf    []byte
}

func NewLogServer(remoteAddr string) *LogServer {
	return &LogServer{remoteAddr: remoteAddr, recvBuf: []byte{0}}
}

func (this *LogServer) Dail() error {
	glog.Infof("Dail to [%s]", this.remoteAddr)
	var (
		err           error
		tcpRemoteAddr *net.TCPAddr
	)

	tcpRemoteAddr, err = net.ResolveTCPAddr("tcp", this.remoteAddr)
	if err != nil {
		glog.Errorf("Resovle Remote TcpAddr [%s] [%s]", this.remoteAddr, err.Error())
		return err
	}
	this.conn, err = net.DialTCP("tcp", nil, tcpRemoteAddr)
	if err != nil {
		glog.Errorf("Dail [%s] [%s]", this.remoteAddr, err.Error())
		return err
	}
	glog.Infof("Dail to [%s] ok", this.remoteAddr)
	return nil
}
func (this *LogServer) Recv() error {
	_, err := this.conn.Read(this.recvBuf)
	return err
}

// func (this *LogServer) Reciver(onHandle func([]byte), onClose func(s *LogServer)) {
// 	defer this.conn.Close()

// 	header := make([]byte, HEADER_SIZE)
// 	buf := make([]byte, PAYLOAD_MAX)

// 	for {
// 		// header
// 		n, err := io.ReadFull(this.conn, header)
// 		if n == 0 && err == io.EOF {
// 			glog.Errorf("[EOF] %v", this.remoteAddr)
// 			break
// 		} else if err != nil {
// 			glog.Errorf("[%s] error receiving header: %s", this.remoteAddr, err.Error())
// 			break
// 		}
// 		size := binary.LittleEndian.Uint32(header)
// 		if size > PAYLOAD_MAX {
// 			glog.Errorf("[%s] overload the max[%d]>[%d]", this.remoteAddr, size, PAYLOAD_MAX)
// 			break
// 		}

// 		data := buf[:size]
// 		n, err = io.ReadFull(this.conn, data)
// 		if n == 0 && err == io.EOF {
// 			glog.Errorf("[EOF] %v", this.remoteAddr)
// 			break
// 		} else if err != nil {
// 			glog.Errorf("[%s] error receiving [%s]", this.remoteAddr, err.Error())
// 			break
// 		}
// 		onHandle(data)
// 	}
// 	onClose(this)
// }

func (this *LogServer) Send(msg []byte) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(msg)))
	binary.Write(buf, binary.LittleEndian, msg)
	_, err := this.conn.Write(buf.Bytes())
	return err
}
