package installservd

import (
	"fmt"

	"github.com/OpenFlowLabs/openindiana-autoinstaller/installd"
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
	Name            string                         `yaml:"name" json:"name"`
	Kernel          *Asset                         `yaml:"-" json:"kernel"`
	KernelName      string                         `yaml:"kernel" json:"-"`
	BootArchive     *Asset                         `yaml:"-" json:"boot_archive"`
	BootArchiveName string                         `yaml:"boot_archive" json:"-"`
	BootArgs        map[string]string              `yaml:"args" json:"args"`
	Config          *installd.InstallConfiguration `yaml:"config" json:"config"`
	Templates       []*Asset                       `yaml:"-" json:"templates"`
	TemplateNames   []string                       `yaml:"templates" json:"-"`
	OS              OSType                         `yaml:"os" json:"os"`
}

func (i *Installservd) AddProfile(profile Profile) (err error) {
	var replyReal string
	//Some Sanity Checks.
	if profile.Config == nil && profile.TemplateNames == nil {
		replyReal = "Either config or a Template must be set in the Profile"
		return fmt.Errorf("error: %s", replyReal)
	}

	if profile.OS == Illumos && profile.Config == nil {
		replyReal = "Illumos requires an Installation config"
		return fmt.Errorf("error: %s", replyReal)
	}

	//Check if we have the Assets named in the Profile.
	if !hasAssets(profile.TemplateNames...) ||
		!hasAssets(profile.BootArchiveName) ||
		!hasAssets(profile.KernelName) {
		replyReal = "Missing Assets Requested by this Project. Please Add Kernel, Boot_Archive and Template as assset beforehand."
		return fmt.Errorf("error: %s", replyReal)
	}

	profile.Kernel = Assets[profile.KernelName]
	profile.BootArchive = Assets[profile.BootArchiveName]
	profile.Templates = getAssets(profile.TemplateNames...)
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
