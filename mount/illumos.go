// +build illumos

package mount

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"errors"

	"github.com/sirupsen/logrus"
)

const (
	lofiadmBin string = "/usr/sbin/lofiadm"
	mountBin   string = "/usr/sbin/mount"
)

func LoopDevice(fstype string, mountpoint string, file string) error {
	lofibuff, lofierr, err := lofiExec([]string{"-a", file})
	if err != nil {
		if !strings.Contains(lofierr, "Device busy") {
			return errors.New(strings.TrimSpace(lofierr))
		}
		//If we get Device Busy then we are already available in lofi. We do not need to fail here it is enough to fail
		//while mounting
		//Still we need to find out the device name
		lofibuff, lofierr, err = lofiExec([]string{file})
	}
	var mountbuff, mounterr bytes.Buffer
	mount := exec.Command(mountBin, fmt.Sprintf("-F%s", fstype), "-o", "ro", lofibuff, mountpoint)
	logrus.Trace(mount.Path, mount.Args)
	mount.Stdout = &mountbuff
	mount.Stderr = &mounterr
	if err := mount.Run(); err != nil {
		return errors.New(strings.TrimSpace(mounterr.String()))
	}
	return nil
}

//TODO func IsMounted(device, path) bool

func lofiExec(args []string) (out string, errout string, err error) {
	lofiadm := exec.Command(lofiadmBin, args...)
	var lofibuff, lofierr bytes.Buffer
	lofiadm.Stdout = &lofibuff
	lofiadm.Stderr = &lofierr
	logrus.Trace(lofiadm.Path, lofiadm.Args)
	err = lofiadm.Run()
	out = strings.TrimSpace(lofibuff.String())
	errout = strings.TrimSpace(lofierr.String())
	return
}
