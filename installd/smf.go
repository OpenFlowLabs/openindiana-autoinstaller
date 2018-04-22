package installd

import (
	"os"
	"path/filepath"

	"bytes"
	"text/template"

	"io/ioutil"

	"git.wegmueller.it/toasterson/glog"
)

/*
Profiles that need to be hooked up for configuration
/etc/svc/profile/name_service.xml -> /etc/svc/profile/ns_dns.xml
/etc/svc/profile/generic.xml -> /etc/svc/profile/generic_limited_net.xml
/etc/svc/profile/platform.xml -> /etc/svc/profile/platform_none.xml
/etc/svc/profile/site.xml custom with sysding
*/

var defaultProfileFiles = map[string]string{
	"/etc/svc/profile/name_service.xml":   "ns_dns.xml",
	"/etc/svc/profile/generic.xml":        "generic_limited_net.xml",
	"/etc/svc/profile/platform.xml":       "platform_none.xml",
	"/etc/svc/profile/inetd_services.xml": "inetd_generic.xml",
}

func hookUpServiceManifests(conf *InstallConfiguration, rootDir string, noop bool) error {
	os.Chdir(filepath.Join(rootDir, "etc/svc/profile"))
	for pFile, target := range defaultProfileFiles {
		if noop {
			glog.Infof("Linking %s -> %s", pFile, target)
			continue
		}
		if err := os.Symlink(target, pFile); err != nil {
			if !os.IsExist(err) {
				glog.Errf("Could not create link %s -> %s: %s", pFile, target, err)
				continue
			}
			return err
		}
	}
	os.Chdir("/tmp")
	tplSiteXML, err := template.New("SiteXML").Parse(siteTemplate)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = tplSiteXML.Execute(&out, conf)
	if err != nil {
		return err
	}
	if noop {
		glog.Infof("Would write site.xml %s", out.String())
		return nil
	}
	glog.Infof("Writing site.xml %s", out.String())
	return ioutil.WriteFile(filepath.Join(rootDir, "etc/svc/profile/site.xml"), out.Bytes(), 0644)
}

const siteTemplate = `<?xml version='1.0'?>
<!DOCTYPE service_bundle SYSTEM "/usr/share/lib/xml/dtd/service_bundle.dtd.1">
<service_bundle type='profile' name='installd_profile'>
    <service name='system/keymap' version='1' type='service'>
        <instance name='default'>
            <property_group name='keymap' type='system'>
                <propval name='layout' type='astring' value='{{.Keymap}}' />
            </property_group>
        </instance>
    </service>
    <service name='system/sysding' version='1' type='service'>
       <instance name='system' enabled='true' />
    </service>
</service_bundle>`
