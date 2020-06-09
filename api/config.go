package api

import (
	"fmt"

	"github.com/zoumo/goset"
)

const (
	localHTTPPort = 1087
)

const (
	NetworkHTTP      = "http"
	NetworkWebsocket = "ws"
	NetworkTCP       = "tcp"
	NetworkKCP       = "kcp"
)

func VmessTag(network string, suffix ...interface{}) string {

	tag := fmt.Sprintf("vmess-%v", network)
	for _, s := range suffix {
		tag += fmt.Sprintf("-%v", s)
	}
	return tag
}

func Networks() []string {
	return []string{
		NetworkHTTP,
		NetworkTCP,
		NetworkKCP,
	}
}

func NewDefaultConfig(user User) Config {
	return Config{
		user: user,
		Log: Log{
			Access:   "/var/log/v2ray/access.log",
			Error:    "/var/log/v2ray/error.log",
			Loglevel: "warning",
		},
		States: States{},
		Policy: Policy{
			System: SystemPolicy{
				StatsInboundUplink:   true,
				StatsInboundDownlink: true,
			},
		},
		Inbounds: []Inbound{},
		Outbounds: []Outbound{
			{
				Tag:      "blocked",
				Protocol: "blackhole",
			},
			{
				Tag:      "freedom",
				Protocol: "freedom",
			},
		},
		Routing: Routing{
			Balancers: []RoutingBalancer{
				// {
				// 	Tag: "vmess-tcp-all",
				// 	Selector: []string{
				// 		"vmess-tcp-",
				// 	},
				// },
				// {
				// 	Tag: "vmess-kcp-all",
				// 	Selector: []string{
				// 		"vmess-kcp-",
				// 	},
				// },
				// {
				// 	Tag: "vmess-http-all",
				// 	Selector: []string{
				// 		"vmess-http-",
				// 	},
				// },
			},
			Rules: []RoutingRule{
				{
					Type:        "field",
					IP:          []string{"geoip:private"},
					OutboundTag: "blocked",
				},
				{
					Type:        "field",
					Domain:      []string{"geosite:category-ads-all"},
					OutboundTag: "blocked",
				},
			},
		},
	}
}

// Config ...
type Config struct {
	Log       Log        `json:"log"`
	States    States     `json:"states"`
	Policy    Policy     `json:"policy"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   Routing    `json:"routing"`
	user      User
}

func (c *Config) InboundVmessNetworks() goset.Set {
	vmessNetwork := goset.NewSet()
	for i := range c.Inbounds {
		if c.Inbounds[i].Protocol != "vmess" {
			continue
		}
		vmessNetwork.Add(c.Inbounds[i].StreamSettings.Network)
	}
	return vmessNetwork
}

func (c *Config) Fillin() {
	vmessNetwork := goset.NewSet()

	for i := range c.Inbounds {
		if c.Inbounds[i].Protocol != "vmess" {
			continue
		}

		// fill user
		c.Inbounds[i].Settings = InboundSettings{
			Clients:                   []User{c.user},
			DisableInsecureEncryption: true,
		}
		// fill sniffing
		vmessNetwork.Add(c.Inbounds[i].StreamSettings.Network)
		if c.Inbounds[i].StreamSettings.Network == NetworkHTTP {
			if c.Inbounds[i].Sniffing == nil {
				c.Inbounds[i].Sniffing = &SniffingSettings{
					Enable:       true,
					DestOverride: []string{"http", "tls"},
				}
			}
		}
	}

	// append vmess all rules
	vmessNetwork.Range(func(k int, v interface{}) bool {
		network := v.(string)

		name := VmessTag(network, "all")
		selector := VmessTag(network, "")
		c.Routing.Rules = append(c.Routing.Rules, RoutingRule{
			Type:        "field",
			InboundTag:  []string{name},
			BalancerTag: name,
		})
		c.Routing.Balancers = append(c.Routing.Balancers, RoutingBalancer{
			Tag: name,
			Selector: []string{
				selector,
			},
		})
		return true
	})

}

func (c *Config) AppendLocalHTTPRule() {
	c.Inbounds = append(c.Inbounds, Inbound{
		Tag:      "localIn",
		Listen:   "127.0.0.1",
		Port:     localHTTPPort,
		Protocol: "http",
	})
	c.Routing.Rules = append(c.Routing.Rules, RoutingRule{
		Type:        "field",
		InboundTag:  []string{"localIn"},
		BalancerTag: "vmess-tcp-all",
	})
}

type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}

type States struct{}

type Policy struct {
	System SystemPolicy `json:"system,omitempty" bson:"system,omitempty"`
}

type SystemPolicy struct {
	StatsInboundUplink   bool `json:"statsInboundUplink,omitempty" bson:"statsInboundUplink,omitempty"`
	StatsInboundDownlink bool `json:"statsInboundDownlink,omitempty" bson:"statsInboundDownlink,omitempty"`
}

type User struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	Level    int    `json:"level,omitempty" bson:"level,omitempty"`
	AlterID  int    `json:"alterId,omitempty" bson:"alterID,omitempty"`
	Security string `json:"security,omitempty" bson:"security,omitempty"`
}
