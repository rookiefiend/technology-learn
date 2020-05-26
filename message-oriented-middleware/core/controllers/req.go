package controllers

type RegistryReq struct {
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
