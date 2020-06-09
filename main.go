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
	"github.com/zoumo/v2ray/api"
)

var (
	vmessNetPort = map[string]int{
		// api.NetworkHTTP:      20000,
		api.NetworkWebsocket: 22000,
	}
	domains = []string{"zoumo.buzz", "us.zoumo.buzz"}
)

const (
	tunnelVmessH2Port = 443
	localHTTPPort     = 1087

	queryPath = "/test"
)

func createHTTPConfig() api.StreamConfig {
	return api.StreamConfig{
		Network:  api.NetworkHTTP,
		Security: "tls",
		HTTPSettings: &api.HTTPSettings{
			Path: queryPath,
		},
		TLSSettings: &api.TLSSettings{
			AllowInsecure: true,
		},
	}
}

func createWSConfig() api.StreamConfig {
	return api.StreamConfig{
		Network: api.NetworkWebsocket,
		WSSettings: &api.WSSettings{
			Path: queryPath,
		},
	}
}

func createKCPConfig() api.StreamConfig {
	return api.StreamConfig{
		Network: api.NetworkKCP,
		KCPSettings: &api.KCPSettings{
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

func createVmessInbound(port int, network string, user api.User) api.Inbound {
	user.Security = ""
	inbound := api.Inbound{
		Listen:   "127.0.0.1",
		Tag:      api.VmessTag(network, "all"),
		Port:     port,
		Protocol: "vmess",
	}
	switch network {
	case api.NetworkHTTP:
		inbound.StreamSettings = createHTTPConfig()
	case api.NetworkWebsocket:
		inbound.StreamSettings = createWSConfig()
	case api.NetworkKCP:
		inbound.StreamSettings = createKCPConfig()
	}

	return inbound
}

func createVmessOutboundLocation(loc string, index int, network string, vnext api.OutboundSettingsVnext) api.Outbound {
	outbound := api.Outbound{
		Tag:      api.VmessTag(network, loc, index),
		Protocol: "vmess",
		Settings: api.OutboundSettings{
			OutboundSettingsVnext: []api.OutboundSettingsVnext{vnext},
		},
	}

	switch network {
	case api.NetworkHTTP:
		outbound.StreamSettings = createHTTPConfig()
	case api.NetworkWebsocket:
		outbound.StreamSettings = createWSConfig()
	case api.NetworkKCP:
		outbound.StreamSettings = createKCPConfig()
	}

	return outbound
}

func newDefaultServerConfig(user api.User) api.Config {
	config := api.NewDefaultConfig(user)

	// add inbounds
	// config.Inbounds = append(config.Inbounds, createVmessInbound(vmessH2Port, api.NetworkHTTP, user))
	for net, port := range vmessNetPort {
		config.Inbounds = append(config.Inbounds, createVmessInbound(port, net, user))
	}
	// add outbounds
	for _, network := range config.InboundVmessNetworks().ToStrings() {
		config.Outbounds = append(config.Outbounds, api.Outbound{
			Tag:      fmt.Sprintf("vmess-%v-all", network),
			Protocol: "freedom",
		})
	}
	config.Fillin()
	return config
}

func newDefaultTunnelConfig(vpses []vps, user api.User) api.Config {
	config := api.NewDefaultConfig(user)
	// add inbounds
	for net, port := range vmessNetPort {
		config.Inbounds = append(config.Inbounds, createVmessInbound(port, net, user))
	}
	// add outbounds
	outbounds := make([]api.Outbound, 0)
	locs := goset.NewSet()

	for i, vps := range vpses {
		locs.Add(vps.Location)
		vnext := api.OutboundSettingsVnext{
			Address: vps.IP,
			Port:    tunnelVmessH2Port,
			Users:   []api.User{user},
		}
		for _, network := range config.InboundVmessNetworks().ToStrings() {
			outbounds = append(outbounds,
				createVmessOutboundLocation(vps.Location, i, network, vnext),
			)
		}

	}
	config.Outbounds = append(config.Outbounds, outbounds...)

	// add loc roles
	config.Routing.Rules = append(config.Routing.Rules, []api.RoutingRule{
		{
			Type:        "field",
			IP:          []string{"geoip:cn"},
			OutboundTag: "freedom",
		},
		{
			Type:        "field",
			Domain:      []string{"geosite:cn"},
			OutboundTag: "freedom",
		},
	}...)

	locs.Range(func(i int, elem interface{}) bool {
		loc := elem.(string)
		for _, network := range config.InboundVmessNetworks().ToStrings() {
			networkB, networkR := createRouting(network, loc)
			config.Routing.Balancers = append(config.Routing.Balancers, networkB)
			config.Routing.Rules = append(config.Routing.Rules, networkR)
		}
		return true
	})

	config.Fillin()
	return config
}

func createRouting(network, location string) (api.RoutingBalancer, api.RoutingRule) {
	prefix := fmt.Sprintf("vmess-%v-%v-", network, location)
	balancer := api.RoutingBalancer{
		Tag: prefix + "all",
		Selector: []string{
			prefix,
		},
	}
	rule := api.RoutingRule{
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

func (c *config) buildUser() api.User {
	return api.User{
		ID:       c.User,
		Level:    1,
		AlterID:  64,
		Security: "auto",
	}
}

func createSubscribe(user api.User) []byte {
	createUrl := func(domain, network, path string) api.VmessURL {
		return api.VmessURL{
			Add:  domain,
			Port: strconv.Itoa(443),
			Net:  network,
			PS:   domain,
			V:    "2",
			ID:   user.ID,
			Type: "none",
			AID:  "4",
			TlS:  "tls",
			Path: path,
		}
	}
	var urls []string
	for _, domain := range domains {
		for net := range vmessNetPort {
			url := createUrl(domain, net, queryPath)
			urls = append(urls, url.URL())
		}
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

	config := newDefaultServerConfig(user)
	filename := "./output/config-server.json"
	j, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

	config = newDefaultTunnelConfig(userConfig.Vpses, user)
	filename = "./output/config-tunnel.json"
	j, _ = json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filename, j, 0755)

	vpses := []vps{}
	vpses = append(vpses, userConfig.Tunnels...)
	vpses = append(vpses, userConfig.Vpses...)
	subscribe := createSubscribe(user)
	filename = "./output/subscirbe.txt"
	ioutil.WriteFile(filename, subscribe, 0755)
}
