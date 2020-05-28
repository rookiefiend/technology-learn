package comm

import (
	"context"
	"net"
	"net/http"
	"technology/message-oriented-middleware/conn/types"
	"time"
)

var (
	DefaultReliableTransport = NewReliableTransport()
)

type ReliableTransport struct {
	http.Transport
}

// NewReliableTransport ...
func NewReliableTransport() *ReliableTransport {
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			conn, err = net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			conn = types.NewReliableConn(conn)
			return conn, nil
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   0,
		ExpectContinueTimeout: 0,
	}
	return &ReliableTransport{tr}
}

type ReliableListener struct {
	listener net.Listener
}

func (rl *ReliableListener) Close() error {
	return rl.listener.Close()
}

func (rl *ReliableListener) Addr() net.Addr {
	return rl.listener.Addr()
}

func (rl *ReliableListener) Accept() (net.Conn, error) {
	conn, err := rl.listener.Accept()
	if err != nil {
		return nil, err
	}
	conn = types.NewReliableConn(conn)
	return conn, nil
}

func NewReliableListener(network, addr string) (*ReliableListener, error) {
	var err error
	rl := new(ReliableListener)
	rl.listener, err = net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return rl, nil
}

// ResponseData ...
type ResponseData struct {
	Msg  string      `json:"msg,omitempty"`
	Err  string      `json:"err,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
