package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"technology/message-oriented-middleware/comm"
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
			Transport: comm.NewReliableTransport(),
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

	respData := new(comm.ResponseData)
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
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
