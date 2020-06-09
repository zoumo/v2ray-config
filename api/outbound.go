package api

type Outbound struct {
	Tag            string           `json:"tag,omitempty" bson:"tag,omitempty"`
	Protocol       string           `json:"protocol,omitempty" bson:"protocol,omitempty"`
	Settings       OutboundSettings `json:"settings,omitempty" bson:"settings,omitempty"`
	StreamSettings StreamConfig     `json:"streamSettings,omitempty" bson:"streamSettings,omitempty"`
}

type OutboundSettings struct {
	OutboundSettingsVnext []OutboundSettingsVnext `json:"vnext,omitempty"`
}

type OutboundSettingsVnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}
