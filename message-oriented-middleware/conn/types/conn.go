package types

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxReceiveCacheSize = 1024
	ackTimeoutDuration  = 100 * time.Hour
)

// ReliableConn 可靠连接，确保发送的每个包都被连接的另一端所接收
type ReliableConn struct {
	idCursor         int64
	idCursorMutex    sync.Mutex
	waitAckCh        chan int64
	receiveDataCh    chan []byte
	receiveDataCache *bytes.Buffer
	conn             net.Conn
	stopCh           chan struct{}
	isClose          bool
	isCloseMutex     sync.Mutex
	writeMutex       sync.Mutex
	isWriting        bool
}

// NewReliableConn ...
func NewReliableConn(conn net.Conn) *ReliableConn {
	rc := new(ReliableConn)
	rc.conn = conn
	rc.waitAckCh = make(chan int64)
	rc.receiveDataCh = make(chan []byte, maxReceiveCacheSize)
	rc.receiveDataCache = bytes.NewBuffer(nil)
	rc.stopCh = make(chan struct{})
	go rc.underlyingRead()
	return rc
}

// underlyingRead 从另一端接收数据，并按类型传入不同的位置
func (rc *ReliableConn) underlyingRead() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-rc.stopCh:
			cancel()
			close(rc.receiveDataCh)
			return
		default:
		}

		h, err := ReceiveHeader(ctx, rc.conn)
		if err != nil {
			logrus.Errorf("failed to receive header, error = %v", err)
			rc.Close()
			continue
		}
		dataBytes, err := ReadBytes(ctx, rc.conn, int(h.Length))
		if err != nil {
			logrus.WithField("length", h.Length).Errorf("failed to read data bytes")
			rc.Close()
			continue
		}
		switch h.Tpy {
		case AckPackageType:
			rc.waitAckCh <- h.Ack
		case ReqPackageType:
			ackPkg := NewPackage(0, AckPackageType, h.Id, nil)
			ackPkgBytes, err := ackPkg.MarshalBytes()
			if err != nil {
				logrus.WithField("package", ackPkg).Errorf("failed to marshal ack package, error = %v", err)
				continue
			}
			_, err = rc.conn.Write(ackPkgBytes)
			if err != nil {
				logrus.WithField("package", ackPkg).Errorf("failed to send ack package, error = %v", err)
				continue
			}
			rc.receiveDataCh <- dataBytes
		default:
			logrus.WithField("type", h.Tpy).
				Errorf("accept unknown type package")
			return
		}

	}

}

func (rc *ReliableConn) Read(b []byte) (n int, err error) {
	var (
		rn int
	)
	rn, err = rc.receiveDataCache.Read(b)
	if err != nil && err != io.EOF {
		return rn, err
	}
	if rn == len(b) {
		return rn, nil
	}

	for {
		select {
		case <-rc.stopCh:
			return rn, fmt.Errorf("read from close conn")
		case data, isOpened := <-rc.receiveDataCh:
			if !isOpened {
				return rn, err
			}
			i := 0
			for ; i < len(data) && rn+i < len(b); i++ {
				b[rn+i] = data[i]
			}
			rn = rn + i
			if rn == len(b) {
				rc.receiveDataCache.Write(data[i:])
				return rn, nil
			}
		default:
			if rn == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return rn, nil
		}
	}
}

func (rc *ReliableConn) Write(b []byte) (rn int, err error) {
	rc.writeMutex.Lock()
	defer rc.writeMutex.Unlock()
	pkg := NewPackage(rc.idCursor, ReqPackageType, 0, b)
	rc.idCursor++
	pkgBytes, err := pkg.MarshalBytes()
	if err != nil {
		return 0, err
	}

	n, err := rc.conn.Write(pkgBytes)
	if err != nil {
		return 0, err
	}

	timer := time.NewTimer(ackTimeoutDuration)
	defer timer.Stop()
	for {
		select {
		case <-rc.stopCh:
			return 0, fmt.Errorf("not accept ack package, conn has been close, bytes = %s", b)
		case <-timer.C:
			return 0, fmt.Errorf("wait ack package timeout, bytes = %s", b)
		case ackId := <-rc.waitAckCh:
			if ackId != pkg.Header.Id {
				continue
			}
			return n - HeaderLength, nil
		}
	}
}

func (rc *ReliableConn) Close() error {
	rc.isCloseMutex.Lock()
	defer rc.isCloseMutex.Unlock()
	if rc.isClose {
		return nil
	}
	close(rc.stopCh)
	rc.isClose = true
	rc.conn.Close()
	return nil
}

func (rc *ReliableConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

func (rc *ReliableConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

func (rc *ReliableConn) SetDeadline(t time.Time) error {
	return nil
}

func (rc *ReliableConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (rc *ReliableConn) SetWriteDeadline(t time.Time) error {
	return nil
}
