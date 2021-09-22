// +build illumos

package installd

import "os/exec"

const devfsadmBin = "/usr/sbin/devfsadm"

func runDevfsadm(root string, device_classes []string) (err error) {
	var args []string
	if root != "" {
		args = append(args, "-r", root)
	}
	for _, dev_class := range device_classes {
		args = append(args, "-c", dev_class)
	}
	devfsadm := exec.Command(devfsadmBin, args...)
	if err = devfsadm.Run(); err != nil {
		return
	}
	return
}
