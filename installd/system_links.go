// +build illumos

package installd

import (
	"os"

	"path/filepath"

	"github.com/sirupsen/logrus"
)

type LinkConfig struct {
	Name   string
	Target string
}

var defaultLinks = []LinkConfig{
	{Name: "stderr", Target: "../fd/2"},
	{Name: "stdout", Target: "../fd/1"},
	{Name: "stdin", Target: "../fd/0"},
	{Name: "dld", Target: "../devices/pseudo/dld@0:ctl"},
}

func makeDeviceLinks(rootDir string, links []LinkConfig, noop bool) error {
	links = append(links, defaultLinks...)
	for _, link := range links {
		path := filepath.Join(rootDir, "dev", link.Name)
		if noop {
			logrus.Infof("Would create device link %s -> %s", path, link.Target)
			continue
		}
		if _, osexisterr := os.Lstat(path); osexisterr == nil {
			logrus.Infof("Symlink %s already existing Distribution does not need it ignoring", path)
			continue
		}
		err := os.Symlink(link.Target, path)
		if os.IsExist(err) {
			return err
		}
	}
	return nil
}
