package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

const (
	ReqPackageType PackageType = iota
	AckPackageType
)

const (
	HeaderLength = 8 + 1 + 8 + 8
)

type PackageType int8

type Header struct {
	Id     int64
	Tpy    PackageType
	Ack    int64
	Length int64
}

func (h Header) MarshalBytes() ([]byte, error) {
	bsBuf := bytes.NewBuffer(nil)
	err := binary.Write(bsBuf, binary.LittleEndian, h)
	if err != nil {
		return nil, err
	}
	return bsBuf.Bytes(), nil
}

func (h *Header) UnmarshalBytes(data []byte) error {
	if len(data) != HeaderLength {
		return fmt.Errorf("invalid header bytes")
	}
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, h)
	if err != nil {
		return err
	}
	return nil
}

type Package struct {
	Header Header
	Body   []byte
}

func NewPackage(id int64, tpy PackageType, ack int64, body []byte) Package {
	return Package{
		Header: Header{
			Id:     id,
			Tpy:    tpy,
			Ack:    ack,
			Length: int64(len(body)),
		},
		Body: body,
	}
}

func (p Package) MarshalBytes() ([]byte, error) {
	hBytes, err := p.Header.MarshalBytes()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hBytes)
	if p.Body != nil {
		buf.Write(p.Body)
	}
	return buf.Bytes(), nil
}

func (p *Package) UnmarshalBytes(data []byte) error {
	err := (&p.Header).UnmarshalBytes(data[:HeaderLength])
	if err != nil {
		return err
	}

	p.Body = data[HeaderLength:]
	return nil
}

// ReceiveHeader ...
func ReceiveHeader(ctx context.Context, conn net.Conn) (*Header, error) {
	h := new(Header)
	headerBytes, err := ReadBytes(ctx, conn, HeaderLength)
	if err != nil {
		logrus.WithField("conn", conn).WithField("length", HeaderLength).
			Errorf("failed to read bytes, error = %v", err)
		return h, err
	}

	err = h.UnmarshalBytes(headerBytes)
	if err != nil {
		logrus.Errorf("failed to unmarshal header bytes, error = %v", err)
		return h, err
	}
	return h, nil
}

func ReadBytes(ctx context.Context, conn net.Conn, length int) ([]byte, error) {
	var (
		readLen = 0
		n       int
		err     error
	)
	if length <= 0 {
		return nil, nil
	}
	bodyBytes := make([]byte, length)
	bodyBytesT := bodyBytes
	for {
		select {
		case <-ctx.Done():
			return nil, contextCancelError
		default:

		}
		n, err = conn.Read(bodyBytesT)
		readLen += n
		if err != nil {
			if err == io.EOF && readLen == length {
				break
			}
			logrus.WithField("conn", conn).
				Errorf("failed to read bytes, error = %v", err)
			return nil, err
		}
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
