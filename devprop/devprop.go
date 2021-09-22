package devprop

import (
	"bytes"
	"os/exec"
)

const devpropBin string = "/sbin/devprop"

func GetValue(key string) (value string) {
	cmd := exec.Command(devpropBin, key)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return
	}
	value = string(out.Bytes())
	return
}
