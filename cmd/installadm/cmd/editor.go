package cmd

import (
	"io/ioutil"
	"os"

	"os/exec"

	"github.com/OpenFlowLabs/openindiana-autoinstaller/installservd"
	"gopkg.in/yaml.v2"
)

func editProfileWithEditor(profile *installservd.Profile) error {
	var tmpFile *os.File
	var err error
	editor := os.ExpandEnv("$EDITOR")
	if tmpFile, err = ioutil.TempFile("", "profile_yaml"); err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if err = yaml.NewEncoder(tmpFile).Encode(profile); err != nil {
		return err
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fileName := tmpFile.Name()

	tmpFile.Close()
	if tmpFile, err = os.Open(fileName); err != nil {
		return err
	}

	if err = yaml.NewDecoder(tmpFile).Decode(profile); err != nil {
		return err
	}
	return nil
}
