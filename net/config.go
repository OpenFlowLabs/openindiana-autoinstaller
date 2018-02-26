package net

type NetworkInterface struct {
	Name   string `json:"name"`   //Name the AddrObj or vnic gets
	IPv4   string `json:"ipv4"`   //Network interfaces IP address Allowed are N.N.N.N/NN for static or dhcp for dhcp
	IPv6   string `json:"ipv6"`   //The IP Version 6 Address of this Interface
	Device string `json:"device"` //The Physical device which to use for configuration Either directly e1000g0 for first intel link or vnic0 for vnic
	Type   string `json:"type"`   //Type can be either vnic, etherstub, iface(Default)
}

type Routes struct {
	Name    string `json:"name"`    //Human readable route name
	Match   string `json:"match"`   //The Routing rule. default if nothing is mentioned
	Gateway string `json:"gateway"` //The Gateway to
}

type NetworkSettings struct {
	Routes     []Routes           `json:"routes"`     //The Routes
	Interfaces []NetworkInterface `json:"interfaces"` //The devices Network interfaces
}
