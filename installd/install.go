// +build solaris

package installd

import (
	"fmt"
	"os"
	"path/filepath"

	"git.wegmueller.it/opencloud/installer/bootadm"
	"git.wegmueller.it/opencloud/installer/mount"
	"git.wegmueller.it/opencloud/opencloud/common"
	"git.wegmueller.it/opencloud/opencloud/zfs"
	"git.wegmueller.it/toasterson/glog"
)

const altRootLocation = "/a"
const altMountLocation = "/mnt.install"
const solusrfileName = "solaris.zlib"
const solmediarootfileName = "boot_archive"

func Install(conf InstallConfiguration, noop bool) error {
	if err := createAndMountZpool(&conf, noop); err != nil {
		return err
	}
	if err := createDatasets(&conf, noop); err != nil {
		return err
	}
	if err := installOS(&conf); err != nil {
		return err
	}
	rootDir := altRootLocation
	if conf.InstallType == InstallTypeBootEnv {
		rootDir = GetPathOfBootEnv(conf.BEName)
	}
	if err := makeSystemDirectories(rootDir, []DirConfig{}, noop); err != nil {
		return err
	}
	if err := makeDeviceLinks(rootDir, []LinkConfig{}, noop); err != nil {
		return err
	}
	if noop {
		glog.Infof("Would run devfsadm on /a")
	} else {
		if err := runDevfsadm(rootDir, []string{}); err != nil {
			return err
		}
	}
	bconf := bootadm.BootConfig{Type: bootadm.BootLoaderTypeLoader, RPoolName: conf.RPoolName, BEName: conf.BEName, BootOptions: []string{}}
	if noop {
		glog.Infof("Would apply the following boot config to disk: %v", bconf)
	} else {
		if err := bootadm.CreateBootConfigurationFiles(rootDir, bconf); err != nil {
			return err
		}
		if err := bootadm.UpdateBootArchive(rootDir); err != nil {
			return err
		}
		if err := bootadm.InstallBootLoader(rootDir, conf.RPoolName); err != nil {
			return err
		}
	}
	//Remove SMF Repository to force regeneration of SMF at first boot.
	//TODO Make own smf package which is a bit more powerfull
	if err := os.Remove(fmt.Sprintf("/%s/etc/svc/repository.db", rootDir)); err != nil {
		return err
	}
	if err := fixZfsMountPoints(&conf, noop); err != nil{
		return err
	}
	return nil
}

func fixZfsMountPoints(conf *InstallConfiguration, noop bool) error {
	var err error
	var bootenv *zfs.Dataset
	if noop {
		glog.Infof("Would set canmount=noauto,mountpoint=/ on %s/ROOT/%s", conf.RPoolName, conf.BEName)
		return nil
	}
	if bootenv, err = zfs.OpenDataset(fmt.Sprintf("%s/ROOT/%s", conf.RPoolName, conf.BEName)); err !=nil {
		return err
	}
	if err = bootenv.SetProperty("canmount", "noauto"); err != nil {
		return err
	}
	if err = bootenv.SetProperty("mountpoint", "/"); err != nil {
		return err
	}
	return nil
}

func installOS(conf *InstallConfiguration) (err error) {
	switch conf.MediaType {
	case MediaTypeSolNetBoot:
		//Get the files Needed to /tmp
		if err = HTTPDownload(fmt.Sprintf("%s/%s", conf.MediaURL, solusrfileName), "/tmp"); err != nil {
			return err
		}
		if err = HTTPDownload(fmt.Sprintf("%s/platform/i86pc/%s", conf.MediaURL, solmediarootfileName), "/tmp"); err != nil {
			return err
		}
		return installOSFromMediaFiles("/tmp")
	case MediaTypeSolCDrom:
	case MediaTypeSolUSB:
		//Assume everything needed is located under /.cdrom
		installOSFromMediaFiles("/.cdrom")
	case MediaTypeZImage:
		return common.NotSupportedError("Image installation")
	case MediaTypeACI:
		if err = HTTPDownload(fmt.Sprintf("%s/%s", conf.MediaURL, "image.aci"), "/tmp"); err != nil {
			return err
		}
		//TODO ACI To Disk Writer
	default:
		return common.InvalidConfiguration("MediaType")
	}
	return
}

func installOSFromMediaFiles(saveLocation string) error {
	var err error
	os.Mkdir(altMountLocation, os.ModeDir)
	if err = mount.MountLoopDevice("ufs", altMountLocation, fmt.Sprintf("%s/%s", saveLocation, solmediarootfileName)); err != nil {
		return err
	}
	if err = mount.MountLoopDevice("hsfs", fmt.Sprintf("%s/usr", altMountLocation), fmt.Sprintf("%s/%s", saveLocation, solusrfileName)); err != nil {
		return err
	}
	filelist := []string{
		"bin",
		"boot",
		"kernel",
		"lib",
		"platform",
		"root",
		"sbin",
		"usr",
		"etc",
		"var",
		"zonelib",
	}
	for _, dir := range filelist {
		filepath.Walk(fmt.Sprintf("%s/%s", altMountLocation, dir), walkCopy)
	}
	return nil
}
