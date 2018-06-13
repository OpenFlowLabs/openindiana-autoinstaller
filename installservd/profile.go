package installservd

import (
	"fmt"

	"git.wegmueller.it/opencloud/installer/installd"
)

var Profiles []Profile

type OSType int

const (
	Illumos OSType = iota
	Linux
	GenericUnix
)

const profileFileName = "profiles.json"

type Profile struct {
	Name                string                         `yaml:"name"`
	Kernel              *Asset                         `yaml:"-"`
	KernelName          string                         `yaml:"kernel"`
	BootArchive         *Asset                         `yaml:"-"`
	BootArchiveName     string                         `yaml:"boot_archive"`
	BootArgs            map[string]string              `yaml:"args"`
	InstallConfig       *installd.InstallConfiguration `yaml:"config"`
	InstallTemplate     *Asset                         `yaml:"-"`
	InstallTemplateName string                         `yaml:"template"`
	OS                  OSType                         `yaml:"os"`
}

func (i *Installservd) AddProfile(profile Profile) (err error) {
	var replyReal string
	//Some Sanity Checks.
	if profile.InstallConfig == nil && profile.InstallTemplateName == "" {
		replyReal = "Either config or a Template must be set in the Profile"
		return fmt.Errorf("error: %s", replyReal)
	}

	if profile.OS == Illumos && profile.InstallConfig == nil {
		replyReal = "Illumos requires an Installation config"
		return fmt.Errorf("error: %s", replyReal)
	}

	//Check if we have the Assets named in the Profile.
	if !hasAsset(profile.InstallTemplateName) ||
		!hasAsset(profile.BootArchiveName) ||
		!hasAsset(profile.KernelName) {
		replyReal = "Missing Assets Requested by this Project. Please Add Kernel, Boot_Archive and Template as assset beforehand."
		return fmt.Errorf("error: %s", replyReal)
	}

	profile.Kernel = Assets[profile.KernelName]
	profile.BootArchive = Assets[profile.BootArchiveName]
	profile.InstallTemplate = Assets[profile.InstallTemplateName]
	Profiles = append(Profiles, profile)
	return nil
}

func (i *Installservd) SaveProfilesToDisk() error {
	return saveToDisk(i.ServerHome, profileFileName, &Profiles)
}

func (i *Installservd) LoadProfilesFromDisk() error {
	return loadFromDisk(i.ServerHome, profileFileName, &Profiles, func() {
		Profiles = make([]Profile, 0)
	})
}
