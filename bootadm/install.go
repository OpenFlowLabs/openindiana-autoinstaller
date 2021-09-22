// +build illumos

package bootadm

import (
	"os/exec"
)

const bootadmBin = "/sbin/bootadm"

func InstallBootLoader(rootDir string, pool string) error {
	args := []string{"install-bootloader"}
	if rootDir != "" {
		args = append(args, "-R", rootDir)
	}
	if pool != "" {
		args = append(args, "-P", pool)
	}
	return execBootadm(args)
}

func execBootadm(args []string) error {
	bootadm := exec.Command(bootadmBin, args...)
	return bootadm.Run()
}
