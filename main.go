package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	vmessPort    = 443
	telegramPort = 442
)

// Config ...
type Config struct {
	Log       Log        `json:"log"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   Routing    `json:"routing"`
}

type Log struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}

type Inbound struct {
	Port     int             `json:"port,omitempty"`
	Protocol string          `json:"protocol,omitempty"`
	Tag      string          `json:"tag,omitempty"`
	Settings InboundSettings `json:"settings,omitempty"`
}

type InboundSettings struct {
	Users                     []InboundUser   `json:"users,omitempty"`
	Clients                   []InboundClient `json:"clients,omitempty"`
	DisableInsecureEncryption bool            `json:"disableInsecureEncryption,omitempty"`
}

type InboundUser struct {
	Secret string `json:"secret"`
}

type InboundClient struct {
	ID       string `json:"id"`
	Level    int    `json:"level"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security,omitempty"`
}

type Outbound struct {
	Tag      string           `json:"tag"`
	Protocol string           `json:"protocol"`
	Settings OutboundSettings `json:"settings,omitempty"`
}

type OutboundSettings struct {
	Vnext    []OutboundSettingsVnext `json:"vnext,omitempty"`
	Redirect string                  `json:"redirect,omitempty"`
}

type OutboundSettingsVnext struct {
	Address string         `json:"address"`
	Port    int            `json:"port"`
	Users   []OutboundUser `json:"users"`
}

type OutboundUser struct {
	ID       string `json:"id"`
	Level    int    `json:"level"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security,omitempty"`
}

type Routing struct {
	Rules     []RoutingRules    `json:"rules"`
	Balancers []RoutingBalancer `json:"balancers"`
}

type RoutingRules struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
	IP          []string `json:"ip,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
}

type RoutingBalancer struct {
	Tag      string   `json:"tag"`
	Selector []string `json:"selector"`
}

type VPS struct {
	IP           string `json:"ip,omitempty" bson:"ip,omitempty"`
	VmessPort    int    `json:"vmessPort,omitempty" bson:"vmessPort,omitempty"`
	TelegramPort int    `json:"telegramPort,omitempty" bson:"telegramPort,omitempty"`
}

type UserConfig struct {
	Vpses  []VPS         `json:"vpses,omitempty" bson:"vpses,omitempty"`
	Client InboundClient `json:"client,omitempty" bson:"client,omitempty"`
	user   OutboundUser
}

func (c *UserConfig) setDefault() {
	for i := range c.Vpses {
		if c.Vpses[i].VmessPort == 0 {
			c.Vpses[i].VmessPort = vmessPort
		}
	}

	if c.Client.Level == 0 {
		c.Client.Level = 1
	}

	if c.Client.AlterID == 0 {
		c.Client.AlterID = 64
	}

	c.user.ID = c.Client.ID
	c.user.Level = c.Client.Level
	c.user.AlterID = c.Client.AlterID
	c.user.Security = "auto"
}

func vpnConfig(vpses []VPS, client InboundClient, user OutboundUser) Config {
	c := defaultConfig(vpses, client)
	vmessOut := Outbound{
		Tag:      "vmess-all",
		Protocol: "freedom",
	}
	// teleOut := Outbound{
	// 	Tag:      "telegram-all",
	// 	Protocol: "mtproto",
	// }
	c.Outbounds = append(c.Outbounds, vmessOut)
	return c
}

func tunnelConfig(vpses []VPS, client InboundClient, user OutboundUser) Config {
	c := defaultConfig(vpses, client)

	port := 8001
	for i, vps := range vpses {
		vmessTag := fmt.Sprintf("vmess-%v", i)
		vmessIn := Inbound{
			Tag:      vmessTag,
			Protocol: "vmess",
			Port:     port,
			Settings: InboundSettings{
				Clients: []InboundClient{
					client,
				},
				DisableInsecureEncryption: true,
			},
		}
		vmessOut := Outbound{
			Tag:      vmessTag,
			Protocol: "vmess",
			Settings: OutboundSettings{
				Vnext: []OutboundSettingsVnext{
					{
						Address: vps.IP,
						Port:    vps.VmessPort,
						Users:   []OutboundUser{user},
					},
				},
			},
		}
		vmessRule := RoutingRules{
			Type:        "field",
			InboundTag:  []string{vmessTag},
			OutboundTag: vmessTag,
		}
		// teleOut := Outbound{
		// 	Tag:      fmt.Sprintf("telegram-%v", i),
		// 	Protocol: "freedom",
		// 	Settings: OutboundSettings{
		// 		Redirect: fmt.Sprintf("%v:%v", vps.IP, vps.TelegramPort),
		// 	},
		// }
		c.Inbounds = append(c.Inbounds, vmessIn)
		c.Outbounds = append(c.Outbounds, vmessOut)
		c.Routing.Rules = append(c.Routing.Rules, vmessRule)
		port++
	}

	return c
}

func defaultConfig(vpses []VPS, client InboundClient) Config {

	config := Config{
		Log: Log{
			Access:   "/var/log/v2ray/access.log",
			Error:    "/var/log/v2ray/error.log",
			Loglevel: "warning",
		},
		Inbounds: []Inbound{
			{
				Port:     80,
				Protocol: "vmess",
				Tag:      "freedom",
				Settings: InboundSettings{
					Clients: []InboundClient{
						client,
					},
					DisableInsecureEncryption: true,
				},
			},
			// {
			// 	Port:     telegramPort,
			// 	Protocol: "mtproto",
			// 	Tag:      "telegram-all",
			// 	Settings: InboundSettings{
			// 		Users: []InboundUser{
			// 			{
			// 				Secret: "ecc84dc93748385a8797366369a2b94a",
			// 			},
			// 		},
			// 	},
			// },
			{
				Port:     vmessPort,
				Protocol: "vmess",
				Tag:      "vmess-all",
				Settings: InboundSettings{
					Clients: []InboundClient{
						client,
					},
					DisableInsecureEncryption: true,
				},
			},
		},
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
				{
					Tag: "vmess-all",
					Selector: []string{
						"vmess-",
					},
				},
				// {
				// 	Tag: "telegram-all",
				// 	Selector: []string{
				// 		"telegram-",
				// 	},
				// },
			},
			Rules: []RoutingRules{
				{
					Type:        "field",
					IP:          []string{"geoip:private"},
					OutboundTag: "blocked",
				},
				{
					Type:        "field",
					InboundTag:  []string{"freedom"},
					OutboundTag: "freedom",
				},
				// {
				// 	Type:        "field",
				// 	InboundTag:  []string{"telegram-all"},
				// 	BalancerTag: "telegram-all",
				// },
				{
					Type:        "field",
					InboundTag:  []string{"vmess-all"},
					BalancerTag: "vmess-all",
				},
			},
		},
	}
	return config
}

func main() {
	file, err := ioutil.ReadFile("./config.json")
	os.Mkdir("output", 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	var userConfig UserConfig
	json.Unmarshal(file, &userConfig)
	userConfig.setDefault()

	fmt.Println(userConfig)

	config := tunnelConfig(userConfig.Vpses, userConfig.Client, userConfig.user)
	filename := "./output/config-tunnel.json"
	j, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

	config = vpnConfig(userConfig.Vpses, userConfig.Client, userConfig.user)
	filename = "./output/config-server.json"
	j, _ = json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

}
