package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

const (
	schemaTag   int64 = 289234
	maxBodySize       = 2048
)

type Package struct {
	Header
	Body
}

func (pkg Package) MarshalBytes() ([]byte, error) {
	bData, err := pkg.Body.MarshalBytes()
	if err != nil {
		return nil, err
	}

	pkg.Header.Length = int64(len(bData))
	hData, err := pkg.Header.MarshalBytes()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(bData)
	buf.Write(hData)
	return buf.Bytes(), nil
}

const (
	SimpleCallPkgType = iota
	AckPkgType
)

const (
	HeaderLength = 8 + 8 + 8 + 1 + 8
)

type PkgType int8

type Header struct {
	Tag    int64   `json:"tag,omitempty"`
	Id     int64   `json:"id,omitempty"`
	Ack    int64   `json:"ack,omitempty"`
	Tpy    PkgType `json:"tpy,omitempty"`
	Length int64   `json:"length,omitempty"`
}

func (h Header) MarshalBytes() ([]byte, error) {
	var err error
	bsBuf := bytes.NewBuffer(nil)
	err = binary.Write(bsBuf, binary.LittleEndian, h)
	if err != nil {
		return nil, err
	}
	return bsBuf.Bytes(), nil
}

func (h *Header) UnmarshalBytes(data []byte) error {
	if len(data) != HeaderLength {
		return fmt.Errorf("invalid header bytes")
	}
	var (
		err error
	)
	reader := bytes.NewReader(data)
	err = binary.Read(reader, binary.LittleEndian, &h)
	if err != nil {
		return err
	}
	return nil
}

type Body struct {
	Key  string `json:"key,omitempty"`
	Data string `json:"data,omitempty"`
}

func (b Body) MarshalBytes() ([]byte, error) {
	return json.Marshal(b)
}

func (b *Body) UnmarshalBytes(data []byte) error {
	return json.Unmarshal(data, b)
}

// NewPackage ...
func NewPackage(key string, id int64, ack int64, v interface{}) (Package, error) {
	vJSON, err := json.Marshal(v)
	if err != nil {
		return Package{}, err
	}

	return Package{
		Header{
			Tag:    schemaTag,
			Tpy:    SimpleCallPkgType,
			Id:     id,
			Ack:    ack,
			Length: 0,
		},
		Body{
			Key:  key,
			Data: string(vJSON),
		},
	}, nil
}

// SendPackage  ...
func SendPackage(conn net.Conn, pkg Package) error {
	pkgBytes, err := pkg.MarshalBytes()
	if err != nil {
		return err
	}
	var n int
	for {
		n, err = conn.Write(pkgBytes)
		if err != nil {
			return err
		}
		if n == len(pkgBytes) {
			break
		}
		pkgBytes = pkgBytes[:n]
	}
	return nil
}

// ReceiveHeader ...
func ReceiveHeader(ctx context.Context, conn net.Conn) (Header, error) {
	h := Header{}
	headerBytes, err := ReadBytes(ctx, conn, HeaderLength)
	if err != nil {
		logrus.WithField("conn", conn).WithField("length", HeaderLength).
			Errorf("failed to read bytes, error = %v", err)
		return h, err
	}

	err = (&h).UnmarshalBytes(headerBytes)
	if err != nil {
		logrus.Errorf("failed to unmarshal header bytes, error = %v", err)
		return h, err
	}
	if h.Tag != schemaTag {
		return h, fmt.Errorf("unknown package header")
	}

	return h, nil
}

func ReceiveBody(ctx context.Context, conn net.Conn, length int) (Body, error) {
	var (
		b   Body
		err error
	)
	bodyBytes, err := ReadBytes(ctx, conn, length)
	if err != nil {
		logrus.WithField("conn", conn).WithField("length", length).
			Errorf("failed to read bytes, error = %v", err)
		return b, err
	}

	err = (&b).UnmarshalBytes(bodyBytes)
	if err != nil {
		return Body{}, err
	}
	return b, nil
}

func ReadBytes(ctx context.Context, conn net.Conn, length int) ([]byte, error) {
	var (
		readLen = 0
		n       int
		err     error
	)
	bodyBytes := make([]byte, length)
	bodyBytesT := bodyBytes
	for {
		select {
		case <-ctx.Done():
			return nil, contextCancelError
		default:

		}
		n, err = conn.Read(bodyBytesT)
		if err != nil && err != io.EOF {
			logrus.WithField("conn", conn).
				Errorf("failed to read bytes, error = %v", err)
			return nil, err
		}
		readLen += n
		if readLen == length {
			break
		}
		bodyBytesT = bodyBytesT[n:]
	}
	return bodyBytes, nil
}

var (
	contextCancelError = fmt.Errorf("context has benn cancel")
)
