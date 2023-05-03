package internal

type ConfigTests struct {
	Protocol string   `json:"protocol"`
	Provider string   `json:"provider"`
	Features []string `json:"features"`
}

type ConfigHardware struct {
	Devices []struct {
		Name     string `json:"name"`
		Vendor   string `json:"vendor"`
		Model    string `json:"model"`
		BmcHost  string `json:"bmcHost" yaml:"bmcHost"` // yaml struct tags defined because the library requires it
		BmcUser  string `json:"bmcUser" yaml:"bmcUser"`
		BmcPass  string `json:"bmcPass" yaml:"bmcPass"`
		IpmiPort string `json:"ipmiPort" yaml:"ipmiPort"`
	} `json:"devices"`
}
