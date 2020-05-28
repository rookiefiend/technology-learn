package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"technology/message-oriented-middleware/comm"

	"github.com/sirupsen/logrus"
)

type Client struct {
	schema     string
	addr       string
	httpClient http.Client
}

func NewClient(addr string) *Client {
	return &Client{
		schema: "http",
		addr:   addr,
		httpClient: http.Client{
			Timeout:   0,
			Transport: comm.DefaultReliableTransport,
		},
	}
}

// RegistryDestName ...
func (c Client) RegistryDestName(req RegistryDestNameReq) error {
	url := fmt.Sprintf("%s://%s/registry", c.schema, c.addr)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("failed to read all body from registry resp, error = %v", err)
		return err
	}
	respData := new(comm.ResponseData)
	err = json.Unmarshal(respBody, respData)
	if err != nil {
		logrus.Errorf("failed to unmarshal resp Body [%s] to respData %#v, error = %v", respBody, respData, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", respData.Err)
	}

	return nil
}

// Consume ...
func (c Client) Consume(req ConsumeReq) (*ConsumeResp, error) {
	consumeResp := &ConsumeResp{}
	url := fmt.Sprintf("%s://%s/consume", c.schema, c.addr)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return consumeResp, err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return consumeResp, err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return consumeResp, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("failed to read all body from consume resp, error = %v", err)
		return nil, err
	}

	respData := new(comm.ResponseData)
	respData.Data = consumeResp
	err = json.Unmarshal(respBody, respData)
	if err != nil {
		logrus.Errorf("failed to unmarshal resp Body [%s] to respData %#v, error = %v", respBody, respData, err)
		return consumeResp, err
	}

	if resp.StatusCode != http.StatusOK {
		return consumeResp, fmt.Errorf("%s", respData.Err)
	}

	consumeResp, ok := respData.Data.(*ConsumeResp)
	if !ok {
		return consumeResp, fmt.Errorf("failed to assert resp data %T to ConsumeResp", respData.Data)
	}

	return consumeResp, nil
}

func (c Client) Product(req ProductReq) error {
	url := fmt.Sprintf("%s://%s/product", c.schema, c.addr)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("failed to read all body from product resp, error = %v", err)
		return err
	}

	respData := new(comm.ResponseData)
	err = json.Unmarshal(respBody, respData)
	if err != nil {
		logrus.Errorf("failed to unmarshal resp body [%s] to respData %#v, error = %v", respBody, respData, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", respData.Err)
	}

	return nil
}

type RegistryDestNameReq struct {
	DestName string `json:"destName,omitempty"`
}

type ConsumeReq struct {
	DestName string `json:"destName,omitempty"`
}

type ConsumeResp struct {
	Msg string `json:"msg,omitempty"`
}

type ProductReq struct {
	DestName string `json:"destName,omitempty"`
	Msg      string `json:"msg,omitempty"`
}
