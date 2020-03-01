package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/zoumo/goset"
)

const (
	tcePort       = 8081
	kcpPort       = 28081
	localHTTPPort = 1087
)

// Config ...
type Config struct {
	Log       Log        `json:"log"`
	States    States     `json:"states"`
	Policy    Policy     `json:"policy"`
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   Routing    `json:"routing"`
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

type Inbound struct {
	Listen         string          `json:"listen,omitempty" bson:"listen,omitempty"`
	Protocol       string          `json:"protocol,omitempty" bson:"protocol,omitempty"`
	Port           int             `json:"port,omitempty" bson:"port,omitempty"`
	Tag            string          `json:"tag,omitempty" bson:"tag,omitempty"`
	Settings       InboundSettings `json:"settings,omitempty" bson:"settings,omitempty"`
	StreamSettings StreamConfig    `json:"streamSettings,omitempty" bson:"streamSetting,omitempty"`
}

type InboundSettings struct {
	Clients                   []User `json:"clients,omitempty"`
	DisableInsecureEncryption bool   `json:"disableInsecureEncryption,omitempty"`
}

type User struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	Level    int    `json:"level,omitempty" bson:"level,omitempty"`
	AlterID  int    `json:"alterId,omitempty" bson:"alterID,omitempty"`
	Security string `json:"security,omitempty" bson:"security,omitempty"`
}

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

type StreamConfig struct {
	Network     string     `json:"network,omitempty" bson:"network,omitempty"`
	KCPSettings *KCPConfig `json:"kcpSettings,omitempty" bson:"kcpSettings,omitempty"`
}

type KCPConfig struct {
	Mtu             uint32          `json:"mtu,omitempty" bson:"mtu,omitempty"`
	Tti             uint32          `json:"tti,omitempty" bson:"tti,omitempty"`
	UpCap           uint32          `json:"uplinkCapacity,omitempty" bson:"upCap,omitempty"`
	DownCap         uint32          `json:"downlinkCapacity,omitempty" bson:"downCap,omitempty"`
	Congestion      bool            `json:"congestion,omitempty" bson:"congestion,omitempty"`
	ReadBufferSize  uint32          `json:"readBufferSize,omitempty" bson:"readBufferSize,omitempty"`
	WriteBufferSize uint32          `json:"writeBufferSize,omitempty" bson:"writeBufferSize,omitempty"`
	HeaderConfig    json.RawMessage `json:"header,omitempty" bson:"headerConfig,omitempty"`
}

type Routing struct {
	Rules     []RoutingRule     `json:"rules"`
	Balancers []RoutingBalancer `json:"balancers"`
}

type RoutingRule struct {
	Type        string   `json:"type,omitempty" bson:"type,omitempty"`
	Network     string   `json:"network,omitempty" bson:"network,omitempty"`
	IP          []string `json:"ip,omitempty" bson:"ip,omitempty"`
	Domain      []string `json:"domain,omitempty" bson:"domain,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty" bson:"inboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty" bson:"balancerTag,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty" bson:"outboundTag,omitempty"`
}

type RoutingBalancer struct {
	Tag      string   `json:"tag"`
	Selector []string `json:"selector"`
}

func NewConfig(user User) Config {
	return Config{
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
		Inbounds: []Inbound{
			createVmessInbound(tcePort, "tcp", user),
			createVmessInbound(kcpPort, "kcp", user),
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
					Tag: "vmess-tcp-all",
					Selector: []string{
						"vmess-tcp-",
					},
				},
				{
					Tag: "vmess-kcp-all",
					Selector: []string{
						"vmess-kcp-",
					},
				},
			},
			Rules: []RoutingRule{
				{
					Type:        "field",
					IP:          []string{"geoip:private"},
					OutboundTag: "blocked",
				},
				{
					Type: "field",
					Domain: []string{
						"geosite:category-ads-all",
					},
					OutboundTag: "blocked",
				},
			},
		},
	}
}

func (c *Config) appendVmessAllRules() {
	rules := []RoutingRule{
		{
			Type:        "field",
			InboundTag:  []string{"vmess-tcp-all"},
			BalancerTag: "vmess-tcp-all",
		},
		{
			Type:        "field",
			InboundTag:  []string{"vmess-kcp-all"},
			BalancerTag: "vmess-kcp-all",
		},
	}
	c.Routing.Rules = append(c.Routing.Rules, rules...)
}

func (c *Config) appendLocalHTTPRule() {
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

func createKCPConfig() StreamConfig {
	return StreamConfig{
		Network: "kcp",
		KCPSettings: &KCPConfig{
			Mtu:             1350,
			Tti:             20,
			UpCap:           20,
			DownCap:         100,
			Congestion:      true,
			ReadBufferSize:  2,
			WriteBufferSize: 2,
			HeaderConfig:    json.RawMessage(`{"type": "wechat-video"}`),
		},
	}
}

func createVmessInbound(port int, network string, user User) Inbound {
	user.Security = ""
	tag := fmt.Sprintf("vmess-%v-all", network)
	inbound := Inbound{
		Tag:      tag,
		Port:     port,
		Protocol: "vmess",
		Settings: InboundSettings{
			Clients: []User{
				user,
			},
			DisableInsecureEncryption: true,
		},
	}
	if network == "kcp" {
		inbound.StreamSettings = createKCPConfig()
	}

	return inbound
}

func createVmessOutboundLocation(loc string, index int, network string, vnext OutboundSettingsVnext) Outbound {
	tag := fmt.Sprintf("vmess-%v-%v-%v", network, loc, index)

	outbound := Outbound{
		Tag:      tag,
		Protocol: "vmess",
		Settings: OutboundSettings{
			OutboundSettingsVnext: []OutboundSettingsVnext{vnext},
		},
	}

	if network == "kcp" {
		outbound.StreamSettings = createKCPConfig()
	}

	return outbound
}

func defaultServerConfig(user User) Config {
	config := NewConfig(user)
	outbounds := []Outbound{
		{
			Tag:      "vmess-tcp-all",
			Protocol: "freedom",
		},
		{
			Tag:      "vmess-kcp-all",
			Protocol: "freedom",
		},
	}
	config.Outbounds = append(config.Outbounds, outbounds...)
	config.appendVmessAllRules()
	config.appendLocalHTTPRule()
	return config
}

func defaultTunnelConfig(vpses []vps, user User) Config {
	config := NewConfig(user)

	outbounds := make([]Outbound, 0)
	locs := goset.NewSet()
	for i, vps := range vpses {
		locs.Add(vps.Location)
		vnext := OutboundSettingsVnext{
			Address: vps.IP,
			Port:    tcePort,
			Users:   []User{user},
		}
		outbounds = append(outbounds,
			createVmessOutboundLocation(vps.Location, i, "tcp", vnext),
			createVmessOutboundLocation(vps.Location, i, "kcp", vnext),
		)
	}
	config.Outbounds = append(config.Outbounds, outbounds...)
	config.Routing.Rules = append(config.Routing.Rules,
		RoutingRule{
			Type: "field",
			IP: []string{
				"geoip:cn",
			},
			OutboundTag: "freedom",
		},
		RoutingRule{
			Type: "field",
			Domain: []string{
				"geosite:cn",
			},
			OutboundTag: "freedom",
		},
	)

	locs.Range(func(i int, elem interface{}) bool {
		loc := elem.(string)
		tcpB, tcpR := createRouting("tcp", loc)
		kcpB, kcpR := createRouting("kcp", loc)
		config.Routing.Balancers = append(config.Routing.Balancers, tcpB, kcpB)
		config.Routing.Rules = append(config.Routing.Rules, tcpR, kcpR)
		return true
	})

	config.appendVmessAllRules()
	config.appendLocalHTTPRule()
	return config
}

func createRouting(network, location string) (RoutingBalancer, RoutingRule) {
	prefix := fmt.Sprintf("vmess-%v-%v-", network, location)
	balancer := RoutingBalancer{
		Tag: prefix + "all",
		Selector: []string{
			prefix,
		},
	}
	rule := RoutingRule{
		Type: "field",
		IP: []string{
			"geoip:" + location,
		},
		Network:     network,
		BalancerTag: prefix + "all",
	}

	return balancer, rule
}

type config struct {
	Tunnels []vps  `json:"tunnels,omitempty" bson:"tunnels,omitempty"`
	Vpses   []vps  `json:"vpses,omitempty" bson:"vpses,omitempty"`
	User    string `json:"user,omitempty" bson:"user,omitempty"`
}

type vps struct {
	IP       string `json:"ip,omitempty" bson:"ip,omitempty"`
	Location string `json:"location,omitempty" bson:"location,omitempty"`
	Cloud    string `json:"cloud,omitempty" bson:"cloud,omitempty"`
}

func (c *config) buildUser() User {
	return User{
		ID:       c.User,
		Level:    1,
		AlterID:  64,
		Security: "auto",
	}
}

type vmessURL struct {
	V    string `json:"v"`
	Add  string `json:"add"`
	Port string `json:"port"`
	PS   string `json:"ps"`
	Net  string `json:"net"`
	ID   string `json:"id"`
	AID  string `json:"aid"`
	TlS  string `json:"tls"`
	Host string `json:"host"`
	Type string `json:"type"`
	Path string `json:"path"`
}

func generateSubscribe(vpses []vps, user User) []byte {
	// vmess://{"port":"443","ps":"aliyun","tls":"none","id":"b2a40065-715c-4eca-a691-6c7b10f4c45e","aid":"4","v":"2","host":"","type":"none","path":"","net":"tcp","add":"115.28.241.43"}
	// vmess://{"port":"443","ps":"aliyun-kcp","tls":"none","id":"b2a40065-715c-4eca-a691-6c7b10f4c45e","aid":"4","v":"2","host":"","type":"none","path":"","net":"kcp","add":"115.28.241.43"}

	createURL := func(vps vps, net string, port int) vmessURL {
		return vmessURL{
			Add:  vps.IP,
			Port: strconv.Itoa(port),
			Net:  net,
			PS:   fmt.Sprintf("%v-%v-%v", vps.Cloud, vps.Location, net),
			V:    "2",
			ID:   user.ID,
			Type: "none",
			AID:  "4",
			TlS:  "none",
		}
	}
	urls := []string{}
	for _, vps := range vpses {
		url := createURL(vps, "tcp", tcePort)
		jsonByte, _ := json.Marshal(url)
		jsonStr := base64.StdEncoding.EncodeToString(jsonByte)
		urls = append(urls, "vmess://"+jsonStr)

		url = createURL(vps, "kcp", kcpPort)
		jsonByte, _ = json.Marshal(url)
		jsonStr = base64.StdEncoding.EncodeToString(jsonByte)
		urls = append(urls, "vmess://"+jsonStr)
	}

	all := []byte(strings.Join(urls, "\n"))
	return []byte(base64.StdEncoding.EncodeToString(all))
}

func main() {
	file, err := ioutil.ReadFile("./config.json")
	os.Mkdir("output", 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	var userConfig config
	json.Unmarshal(file, &userConfig)
	user := userConfig.buildUser()

	config := defaultServerConfig(user)
	filename := "./output/config-server.json"
	j, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

	config = defaultTunnelConfig(userConfig.Vpses, user)
	filename = "./output/config-tunnel.json"
	j, _ = json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

	vpses := []vps{}
	vpses = append(vpses, userConfig.Tunnels...)
	vpses = append(vpses, userConfig.Vpses...)
	subscribe := generateSubscribe(vpses, user)
	filename = "./output/subscirbe.txt"
	ioutil.WriteFile(filename, subscribe, 0755)
}
