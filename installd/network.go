package installd

import "git.wegmueller.it/opencloud/installer/net"

func preconfigureNetwork(conf *InstallConfiguration, noop bool) error {
	net.ConfigureNetworking(&conf.Net, noop)
	return nil
}
