// +build solaris

package installd

import (
	"fmt"
	"os"

	"github.com/toasterson/mozaik/logger"
	"path/filepath"
	"git.wegmueller.it/toasterson/glog"
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

func makeDeviceLinks(rootDir string, links []LinkConfig, noop bool) error{
	links = append(links, defaultLinks...)
	for _, link := range links {
		path := filepath.Join(rootDir, "dev", link.Name)
		if noop {
			glog.Infof("Would create device link %s -> %s", path, link.Target)
		}
		err := os.Symlink(link.Target, path)
		if err != nil {
			return err
		}
	}
	return nil
}
