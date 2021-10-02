package models

type Network struct {
	Name       string `json:"Name"`
	ID         string `json:"Id"`
	Created    string `json:"Created"`
	Scope      string `json:"Scope"`
	Driver     string `json:"Driver"`
	EnableIPv6 bool   `json:"EnableIPv6"`
	Ipam       struct {
		Driver  string `json:"Driver"`
		Options struct {
		} `json:"Options"`
		Config []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
	Internal   bool `json:"Internal"`
	Attachable bool `json:"Attachable"`
	Ingress    bool `json:"Ingress"`
	ConfigFrom struct {
		Network string `json:"Network"`
	} `json:"ConfigFrom"`
	ConfigOnly bool                 `json:"ConfigOnly"`
	Containers map[string]Container `json:"Containers"`
	Options    struct{}             `json:"Options"`
	Labels     struct{}             `json:"Labels"`
}

type NetworkContainer struct {
	Name        string `json:"Name"`
	EndpointID  string `json:"EndpointID"`
	MacAddress  string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}
