package api

type Inbound struct {
	Listen         string            `json:"listen,omitempty" bson:"listen,omitempty"`
	Protocol       string            `json:"protocol,omitempty" bson:"protocol,omitempty"`
	Port           int               `json:"port,omitempty" bson:"port,omitempty"`
	Tag            string            `json:"tag,omitempty" bson:"tag,omitempty"`
	Sniffing       *SniffingSettings `json:"sniffing,omitempty" bson:"sniffing,omitempty"`
	Settings       InboundSettings   `json:"settings,omitempty" bson:"settings,omitempty"`
	StreamSettings StreamConfig      `json:"streamSettings,omitempty" bson:"streamSetting,omitempty"`
}

type SniffingSettings struct {
	Enable       bool     `json:"enable,omitempty" bson:"enable,omitempty"`
	DestOverride []string `json:"destOverride,omitempty" bson:"destOverride,omitempty"`
}

type InboundSettings struct {
	Clients                   []User `json:"clients,omitempty"`
	DisableInsecureEncryption bool   `json:"disableInsecureEncryption,omitempty"`
}
