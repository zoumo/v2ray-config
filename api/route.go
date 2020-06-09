package api


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
