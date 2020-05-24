package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"technology/message-oriented-middleware/rpc/types"

	"github.com/sirupsen/logrus"
)

type SimpleCallFunc func(req []byte) []byte

type Server struct {
	simpleMapMutex sync.RWMutex
	simpleMap      map[string]SimpleCallFunc
	listener       net.Listener
}

// RegisterSimple 注册简单函数
func (s *Server) RegisterSimple(key string, callFunc SimpleCallFunc) error {
	s.simpleMapMutex.Lock()
	s.simpleMap[key] = callFunc
	s.simpleMapMutex.Unlock()
	return nil
}

func (s *Server) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			logrus.WithField("server", s).Errorf("failed to close message server, error = %v", err)
			return err
		}
	}
	return nil
}

func (s *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			s.Close()
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				logrus.Infof("failed to accept conn, error = %v", err)
				return err
			}
			go func() {
				err := s.handleConn(ctx, conn)
				if err != nil {
					logrus.WithField("conn", conn).
						Errorf("failed to handle conn, error = %v", err)
				}
			}()
		}
	}
}

// handleConn 处理每个conn
func (s *Server) handleConn(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return nil
		default:

		}
		header, err := types.ReceiveHeader(ctx, conn)
		if err != nil {
			logrus.Errorf("failed to receive header, error = %v", err)
			return err
		}
		body, err := types.ReceiveBody(ctx, conn, int(header.Length))
		if err != nil {
			logrus.Errorf("failed to receive body, error = %v", err)
			return err
		}
		go s.handlePackage(ctx, types.Package{
			Header: header,
			Body:   body,
		})
	}
}

func (s *Server) handlePackage(ctx context.Context, pkg types.Package) error {
	s.simpleMapMutex.RLock()
	callFunc, ok := s.simpleMap[pkg.Body.Key]
	if !ok {
		err := fmt.Errorf("accept unknow key = %s", pkg.Body.Key)
		logrus.Errorln(err)
		s.simpleMapMutex.RUnlock()
		return err
	}
	s.simpleMapMutex.RUnlock()
	resp := callFunc([]byte(pkg.Body.Data))
	respPkg := types.NewPackage(pkg.Body.Key, pkg.Header.Id)

	ht
}
