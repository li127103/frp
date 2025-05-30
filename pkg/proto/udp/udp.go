package udp

import (
	"encoding/base64"
	"github.com/fatedier/golib/errors"
	"github.com/fatedier/golib/pool"
	"github.com/li127103/frp/pkg/msg"
	"net"
	"sync"
	"time"
)

func NewUDPPacket(buf []byte, laddr, raddr *net.UDPAddr) *msg.UDPPacket {
	return &msg.UDPPacket{
		Content:    base64.StdEncoding.EncodeToString(buf),
		LocalAddr:  laddr,
		RemoteAddr: raddr,
	}
}

func GetContent(m *msg.UDPPacket) (buf []byte, err error) {
	buf, err = base64.StdEncoding.DecodeString(m.Content)
	return
}

func ForwardUserConn(udpConn *net.UDPConn, readCh <-chan *msg.UDPPacket, sendCh chan<- *msg.UDPPacket, bufSize int) {
	// read
	go func() {
		for udpMsg := range readCh {
			buf, err := GetContent(udpMsg)
			if err != nil {
				continue
			}
			_, _ = udpConn.WriteToUDP(buf, udpMsg.RemoteAddr)
		}
	}()

	//write
	buf := pool.GetBuf(bufSize)
	defer pool.PutBuf(buf)

	for {
		n, remoteAddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		// buf[:n] will be encoded to string, so the bytes can be reused
		udpMsg := NewUDPPacket(buf[:n], nil, remoteAddr)

		select {
		case sendCh <- udpMsg:
		default:
		}
	}
}

func Forwarder(dstAddr *net.UDPAddr, readCh <-chan *msg.UDPPacket, sendCh chan<- msg.Message, bufSize int) {
	var mu sync.RWMutex
	udpConnMap := make(map[string]*net.UDPConn)
	// read from dstAddr and write to sendCh
	writerFn := func(raddr *net.UDPAddr, udpConn *net.UDPConn) {
		add := raddr.String()
		defer func() {
			mu.Lock()
			delete(udpConnMap, add)
			mu.Unlock()
			udpConn.Close()
		}()

		buf := pool.GetBuf(bufSize)
		for {
			_ = udpConn.SetReadDeadline(time.Now().Add(30 * time.Second))
			n, _, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			udpMsg := NewUDPPacket(buf[:n], nil, raddr)
			if err = errors.PanicToError(func() {
				select {
				case sendCh <- udpMsg:
				default:
				}
			}); err != nil {
				return
			}
		}
	}

	// read from readCh
	go func() {
		for udpMsg := range readCh {
			buf, err := GetContent(udpMsg)
			if err != nil {
				continue
			}
			mu.Lock()
			udpConn, ok := udpConnMap[udpMsg.RemoteAddr.String()]
			if !ok {
				udpConn, err = net.DialUDP("udp", nil, dstAddr)
				if err != nil {
					mu.Unlock()
					continue
				}
				udpConnMap[udpMsg.RemoteAddr.String()] = udpConn
			}
			mu.Unlock()

			_, err = udpConn.Write(buf)
			if err != nil {
				udpConn.Close()
			}

			if !ok {
				go writerFn(udpMsg.RemoteAddr, udpConn)
			}
		}
	}()

}
