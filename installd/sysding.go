package installd

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"path/filepath"

	"git.wegmueller.it/toasterson/glog"
)

func createSysDingConf(conf *InstallConfiguration, noop bool) error {
	var buffer bytes.Buffer

	if conf.TimeZone != "" {
		buffer.WriteString(fmt.Sprintf("setup_timezone \"%s\"\n", conf.TimeZone))
	}

	if conf.Locale != "" {
		buffer.WriteString(fmt.Sprintf("setup_locale \"%s\"\n", conf.Locale))
	}

	if conf.RootPW != "" {
		buffer.WriteString(fmt.Sprintf("setup_user_password \"root\" \"%s\"\n", conf.RootPW))
	}

	for _, inf := range conf.Net.Interfaces {
		if inf.Type == NetTypeIface {
			if inf.IPv4 != "" {
				if inf.IPv4 == "dhcp" {
					buffer.WriteString(fmt.Sprintf("setup_interface \"%s\" \"%s dhcp\" \"dhcp\"\n", inf.Device, inf.Name))
				} else {
					buffer.WriteString(fmt.Sprintf("setup_interface \"%s\" \"%s v4\" \"%s\"\n", inf.Device, inf.Name, inf.IPv4))
				}
			} else if inf.IPv6 != "" {
				buffer.WriteString(fmt.Sprintf("setup_interface \"%s\" \"%s v6\" \"%s\"\n", inf.Device, inf.Name, inf.IPv6))
			}
		}
	}

	for _, route := range conf.Net.Routes {
		buffer.WriteString(fmt.Sprintf("setup_route \"%s\" \"%s\"\n", route.Match, route.Gateway))
	}

	if conf.Net.DNSDomain != "" {
		buffer.WriteString(fmt.Sprintf("setup_ns_dns \"%s\" \"%s\" \"%s\"\n", conf.Net.DNSDomain, formatAsString(conf.Net.DNSSearch), formatAsString(conf.Net.DNSServers)))
	}

	if noop {
		glog.Infof("Would write the following sysding.cong: %s", buffer.String())
		return nil
	}

	return ioutil.WriteFile(filepath.Join(altRootLocation, "etc/sysding.conf"), buffer.Bytes(), 0400)
}

func formatAsString(servers []string) string {
	var buf bytes.Buffer
	for i, srv := range servers {
		format := " %s"
		if i == 0 {
			format = "%s"
		}
		buf.WriteString(fmt.Sprintf(format, srv))
	}
	return buf.String()
}
