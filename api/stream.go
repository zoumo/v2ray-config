package api

import (
	"encoding/json"
)

type StreamConfig struct {
	Network      string        `json:"network,omitempty" bson:"network,omitempty"`
	Security     string        `json:"security,omitempty" bson:"security,omitempty"`
	HTTPSettings *HTTPSettings `json:"httpSettings,omitempty" bson:"httpSettings,omitempty"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty" bson:"tlsSettings,omitempty"`
	KCPSettings  *KCPSettings  `json:"kcpSettings,omitempty" bson:"kcpSettings,omitempty"`
	WSSettings   *WSSettings   `json:"wsSettings,omitempty" bson:"wsSettings,omitempty"`
}

type WSSettings struct {
	Path    string            `json:"path,omitempty" bson:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty" bson:"headers,omitempty"`
}

type TLSSettings struct {
	AllowInsecure bool `json:"allowInsecure,omitempty" bson:"allowInsecure,omitempty"`
}

type HTTPSettings struct {
	Path string   `json:"path,omitempty" bson:"path,omitempty"`
	Host []string `json:"host,omitempty" bson:"host,omitempty"`
}

type KCPSettings struct {
	Mtu             uint32          `json:"mtu,omitempty" bson:"mtu,omitempty"`
	Tti             uint32          `json:"tti,omitempty" bson:"tti,omitempty"`
	UpCap           uint32          `json:"uplinkCapacity,omitempty" bson:"upCap,omitempty"`
	DownCap         uint32          `json:"downlinkCapacity,omitempty" bson:"downCap,omitempty"`
	Congestion      bool            `json:"congestion,omitempty" bson:"congestion,omitempty"`
	ReadBufferSize  uint32          `json:"readBufferSize,omitempty" bson:"readBufferSize,omitempty"`
	WriteBufferSize uint32          `json:"writeBufferSize,omitempty" bson:"writeBufferSize,omitempty"`
	HeaderConfig    json.RawMessage `json:"header,omitempty" bson:"headerConfig,omitempty"`
}
