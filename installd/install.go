// +build solaris

package installd

import (
	"fmt"
	"os"
	"path/filepath"

	"path"

	"strings"

	"git.wegmueller.it/opencloud/installer/bootadm"
	"git.wegmueller.it/opencloud/installer/mount"
	"git.wegmueller.it/opencloud/opencloud/common"
	"git.wegmueller.it/opencloud/opencloud/gnutar"
	"git.wegmueller.it/opencloud/opencloud/pod"
	"git.wegmueller.it/opencloud/opencloud/zfs"
	"git.wegmueller.it/toasterson/glog"
)

const altRootLocation = "/a"
const altMountLocation = "/mnt.install"
const solusrfileName = "solaris.zlib"
const solmediarootfileName = "boot_archive"

var osfilelist = []string{
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

func Install(conf InstallConfiguration, noop bool) error {
	if err := createAndMountZpool(&conf, noop); err != nil {
		return err
	}
	if err := createDatasets(&conf, noop); err != nil {
		return err
	}
	if err := installOS(&conf, noop); err != nil {
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
	bconf := bootadm.BootConfig{Type: bootadm.BootLoaderTypeLoader, RPoolName: conf.Rpool, BEName: conf.BEName, BootOptions: []string{}}
	if noop {
		glog.Infof("Would apply the following boot config to disk: %v", bconf)
	} else {
		if err := bootadm.CreateBootConfigurationFiles(rootDir, bconf); err != nil {
			return err
		}
		if err := bootadm.UpdateBootArchive(rootDir); err != nil {
			return err
		}
		if err := bootadm.InstallBootLoader(rootDir, conf.Rpool); err != nil {
			return err
		}
	}
	//Remove SMF Repository to force regeneration of SMF at first boot.
	//TODO Make own smf package which is a bit more powerfull
	smfRepo := filepath.Join(rootDir, "etc/svc/repository.db")
	if noop {
		glog.Infof("Would remove smf repo at %s", smfRepo)
	} else {
		glog.Infof("Removing smf repo at %s", smfRepo)
		if err := os.Remove(smfRepo); err != nil {
			glog.Errf("Failure: %s", err)
			return err
		}
		glog.Infof("Success")
	}
	if err := fixZfsMountPoints(&conf, noop); err != nil {
		return err
	}
	return nil
}

func fixZfsMountPoints(conf *InstallConfiguration, noop bool) error {
	var err error
	var bootenv *zfs.Dataset
	if noop {
		glog.Infof("Would set canmount=noauto,mountpoint=/ on %s/ROOT/%s", conf.Rpool, conf.BEName)
		return nil
	}
	glog.Infof("Setting canmount=noauto,mountpoint=/ on %s/ROOT/%s", conf.Rpool, conf.BEName)
	if bootenv, err = zfs.OpenDataset(path.Join(conf.Rpool, "ROOT", conf.BEName)); err != nil {
		glog.Errf("Failure: %s")
		return err
	}
	if err = bootenv.SetProperty("canmount", "noauto"); err != nil {
		glog.Errf("Failure: %s")
		return err
	}
	if err = bootenv.SetProperty("mountpoint", "/"); err != nil {
		glog.Errf("Failure: %s")
		return err
	}
	glog.Infof("Success")
	return nil
}

func installOS(conf *InstallConfiguration, noop bool) (err error) {
	switch conf.InstallImage.Type {
	case MediaTypeSolNetBoot:
		//Get the files Needed to /tmp
		if noop {
			glog.Infof("Would download / image from %s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName)
			glog.Infof("Would Download /usr image from %s/%s", conf.InstallImage.URL, solusrfileName)
		} else {
			glog.Infof("Downloading %s/%s", conf.InstallImage.URL, solusrfileName)
			if err = HTTPDownload(fmt.Sprintf("%s/%s", conf.InstallImage.URL, solusrfileName), "/tmp"); err != nil {
				glog.Errf("Failure: %s", err)
				return err
			}
			glog.Infof("Success")
			glog.Infof("Downloading %s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName)
			if err = HTTPDownload(fmt.Sprintf("%s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName), "/tmp"); err != nil {
				glog.Errf("Failure: %s")
				return err
			}
			glog.Infof("Success")
			return installOSFromMediaFiles("/tmp")
		}
	case MediaTypeSolCDrom:
	case MediaTypeSolUSB:
		//Assume everything needed is located under /.cdrom
		if noop {
			glog.Infof("Would install from CDROM")
			return
		}
		installOSFromMediaFiles("/.cdrom")
	case MediaTypeZImage:
		return common.NotSupportedError("Image installation")
	case MediaTypeACI:
	case strings.ToLower(MediaTypeACI):
		if noop {
			glog.Infof("Would Download %s", conf.InstallImage.URL)
		} else {
			glog.Infof("Downloading %s", conf.InstallImage.URL)
			if err = HTTPDownload(conf.InstallImage.URL, "/tmp"); err != nil {
				glog.Errf("Failure: %s", err)
				return err
			}
			glog.Infof("Success")
		}
		_, fileName := path.Split(conf.InstallImage.URL)
		if aciFile, err := os.Open(filepath.Join("/tmp", fileName)); err != nil {
			glog.Errf("Error Opening File: %s", err)
			return err
		} else {
			if aciRd, err := pod.DecompressingReader(aciFile); err != nil {
				glog.Errf("Error Opening ACI: %s", err)
				return err
			} else {
				//TODO Implement Image checksum
				be, err := zfs.OpenDataset(conf.GetRootDataSetName())
				if err != nil {
					return err
				}
				mntpoint := be.Mountpoint
				be.SetProperty("mountpoint", fmt.Sprintf("%s/rootfs", mntpoint))
				return gnutar.ExtractOneInto("rootfs", mntpoint, aciRd)
			}
		}
	default:
		return common.InvalidConfiguration("MediaType")
	}
	return
}

func installOSFromMediaFiles(saveLocation string) error {
	glog.Infof("Installing OS from images under %s", saveLocation)
	var err error
	os.Mkdir(altMountLocation, os.ModeDir)
	root_image := filepath.Join(saveLocation, solmediarootfileName)
	glog.Infof("Mounting %s at %s", root_image, altMountLocation)
	if err = mount.MountLoopDevice("ufs", altMountLocation, root_image); err != nil {
		glog.Errf("Failure: %s", err)
		return err
	}
	glog.Infof("Success")
	usrImage := filepath.Join(saveLocation, solusrfileName)
	usrMnt := filepath.Join(altMountLocation, "usr")
	glog.Infof("Mounting %s at %s", usrImage, usrMnt)
	if err = mount.MountLoopDevice("hsfs", usrMnt, usrImage); err != nil {
		glog.Errf("Failure: %s", err)
		return err
	}
	glog.Infof("Success")
	glog.Infof("Copying everything recursively from %v", osfilelist)
	for _, dir := range osfilelist {
		if err := filepath.Walk(fmt.Sprintf("%s/%s", altMountLocation, dir), walkCopy); err != nil {
			return err
		}
	}
	return nil
}
