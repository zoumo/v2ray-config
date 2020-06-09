package api

import (
	"encoding/base64"
	"encoding/json"
)

// vmess://{"port":"443","ps":"aliyun","tls":"none","id":"b2a40065-715c-4eca-a691-6c7b10f4c45e","aid":"4","v":"2","host":"","type":"none","path":"","net":"tcp","add":"115.28.241.43"}
// vmess://{"port":"443","ps":"aliyun-kcp","tls":"none","id":"b2a40065-715c-4eca-a691-6c7b10f4c45e","aid":"4","v":"2","host":"","type":"none","path":"","net":"kcp","add":"115.28.241.43"}
type VmessURL struct {
	V    string `json:"v,omitempty" bson:"v,omitempty"`
	Add  string `json:"add,omitempty" bson:"add,omitempty"`
	Port string `json:"port,omitempty" bson:"port,omitempty"`
	PS   string `json:"ps,omitempty" bson:"ps,omitempty"`
	Net  string `json:"net,omitempty" bson:"net,omitempty"`
	ID   string `json:"id,omitempty" bson:"id,omitempty"`
	AID  string `json:"aid,omitempty" bson:"aid,omitempty"`
	TlS  string `json:"tls,omitempty" bson:"tlS,omitempty"`
	Host string `json:"host,omitempty" bson:"host,omitempty"`
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Path string `json:"path,omitempty" bson:"path,omitempty"`
}

func (url VmessURL) URL() string {
	jsonByte, _ := json.Marshal(url)
	jsonStr := base64.StdEncoding.EncodeToString(jsonByte)
	return "vmess://" + jsonStr
}
