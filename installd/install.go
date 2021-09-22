// +build illumos

package installd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"path"

	"io/ioutil"

	"git.wegmueller.it/opencloud/opencloud/zfs"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/bootadm"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/fileutils"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/mount"
	"github.com/sirupsen/logrus"
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
		logrus.Errorf("Failed to create and mount Zpools: %s", err)
		return err
	}

	if err := createDatasets(&conf, noop); err != nil {
		logrus.Errorf("Failed to create required Datasets: %s", err)
		return err
	}

	if err := installOS(&conf, noop); err != nil {
		logrus.Errorf("Failed to unpack the OS: %s", err)
		return err
	}

	rootDir := altRootLocation
	if conf.InstallType == InstallTypeBootEnv {
		rootDir = GetPathOfBootEnv(conf.BEName)
	}

	if err := makeSystemDirectories(rootDir, []DirConfig{}, noop); err != nil {
		logrus.Errorf("Failed to create system directories: %s", err)
		return err
	}

	if err := makeDeviceLinks(rootDir, []LinkConfig{}, noop); err != nil {
		logrus.Errorf("Failed to make device links: %s", err)
		return err
	}

	if err := runDevFsAdm(noop, rootDir); err != nil {
		logrus.Errorf("Failed to run devfsadm on target: %s", err)
		return err
	}

	if err := runBootAdm(conf, noop, rootDir); err != nil {
		logrus.Errorf("Failed to run bootadm: %s", err)
		return err
	}

	if err := createSysDingConf(&conf, noop); err != nil {
		logrus.Errorf("Failed to create sysding.conf: %s", err)
		return err
	}

	if err := writeHostNameToDisk(conf, rootDir); err != nil {
		logrus.Errorf("Failed to write hostname to disk: %s", err)
		return err
	}

	if err := removeSMFRepositoryFile(rootDir, noop); err != nil {
		logrus.Errorf("Failed to remove install SMF repository: %s", err)
		return err
	}

	if err := setupSMFProfiles(&conf, rootDir, noop); err != nil {
		logrus.Errorf("Failed to setup SMF profiles: %s", err)
		return err
	}

	if err := fixZfsMountPoints(&conf, noop); err != nil {
		logrus.Errorf("Failed to setup mountpoints: %s", err)
		return err
	}

	return nil
}

func removeSMFRepositoryFile(rootDir string, noop bool) error {
	//Remove SMF Repository to force regeneration of SMF at first boot.
	//TODO Make own smf package which is a bit more powerfull
	smfRepo := filepath.Join(rootDir, "etc/svc/repository.db")
	if noop {
		logrus.Infof("Would remove smf repo at %s", smfRepo)
	} else {
		logrus.Infof("Removing smf repo at %s", smfRepo)
		if err := os.Remove(smfRepo); err != nil {
			if !os.IsNotExist(err) {
				logrus.Errorf("Failure: %s", err)
				return err
			}
		}
		logrus.Infof("Success")
	}

	return nil
}

func writeHostNameToDisk(conf InstallConfiguration, rootDir string) error {
	return ioutil.WriteFile(filepath.Join(rootDir, "etc/nodename"), []byte(conf.Hostname), 0644)
}

func runBootAdm(conf InstallConfiguration, noop bool, rootDir string) error {
	bconf := bootadm.BootConfig{Type: bootadm.BootLoaderTypeLoader, RPoolName: conf.Rpool, BEName: conf.BEName, BootOptions: []string{}}
	if noop {
		logrus.Infof("Would apply the following boot config to disk: %v", bconf)
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
	return nil
}

func runDevFsAdm(noop bool, rootDir string) error {
	if noop {
		logrus.Infof("Would run devfsadm on /a")
	} else {
		if err := runDevfsadm(rootDir, []string{}); err != nil {
			return err
		}
	}
	return nil
}

func fixZfsMountPoints(conf *InstallConfiguration, noop bool) error {
	var err error
	var bootenv *zfs.Dataset
	if noop {
		logrus.Infof("Would set canmount=noauto,mountpoint=/ on %s/ROOT/%s", conf.Rpool, conf.BEName)
		return nil
	}

	logrus.Infof("Setting canmount=noauto,mountpoint=/ on %s/ROOT/%s", conf.Rpool, conf.BEName)
	if bootenv, err = zfs.OpenDataset(path.Join(conf.Rpool, "ROOT", conf.BEName)); err != nil {
		logrus.Errorf("Failure: %s")
		return err
	}

	if err = bootenv.SetProperty("canmount", "noauto"); err != nil {
		logrus.Errorf("Failure: %s")
		return err
	}

	if mounted, _ := bootenv.IsMounted(); mounted {
		if err = bootenv.Unmount(); err != nil {
			logrus.Errorf("Failure: %s")
			return err
		}
	}

	if err = bootenv.SetProperty("mountpoint", "/"); err != nil {
		logrus.Errorf("Failure: %s")
		return err
	}

	logrus.Infof("Success")
	return nil
}

func installOS(conf *InstallConfiguration, noop bool) error {
	switch conf.InstallImage.Type {
	case MediaTypeSolNetBoot:
		//Get the files Needed to /tmp
		if noop {
			logrus.Infof("Would download / image from %s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName)
			logrus.Infof("Would Download /usr image from %s/%s", conf.InstallImage.URL, solusrfileName)
		} else {
			logrus.Infof("Downloading %s/%s", conf.InstallImage.URL, solusrfileName)
			if err := fileutils.HTTPDownload(fmt.Sprintf("%s/%s", conf.InstallImage.URL, solusrfileName), "/tmp"); err != nil {
				logrus.Errorf("Failure: %s", err)
				return err
			}
			logrus.Infof("Success")
			logrus.Infof("Downloading %s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName)
			if err := fileutils.HTTPDownload(fmt.Sprintf("%s/platform/i86pc/%s", conf.InstallImage.URL, solmediarootfileName), "/tmp"); err != nil {
				logrus.Errorf("Failure: %s")
				return err
			}
			logrus.Infof("Success")
			return installOSFromMediaFiles("/tmp")
		}
	case MediaTypeSolCDrom:
		//Assume everything needed is located under /.cdrom
		if noop {
			logrus.Infof("Would install from CDROM")
			return nil
		}
		return installOSFromMediaFiles("/.cdrom")
	case MediaTypeZImage:
		return NotSupportedError("Image installation")
	case MediaTypeTGZ:
		if noop {
			logrus.Infof("Would install / archive from %s", conf.InstallImage.URL)
		} else {
			logrus.Infof("Downloading %s", conf.InstallImage.URL)
			filePath, err := fileutils.HTTPDownloadTo(conf.InstallImage.URL, "/tmp")
			if err != nil {
				logrus.Errorf("Failure: %s", err)
				return err
			}
			logrus.Infof("Success")
			return installOSFromTGZArchive(filePath)
		}
	default:
		return InvalidConfiguration("MediaType")
	}

	return InvalidConfiguration("MediaType")
}

func installOSFromTGZArchive(tgzArchive string) error {
	logrus.Infof("Installing OS from archive %s", tgzArchive)
	tarBin, err := exec.LookPath("gtar")
	if err != nil {
		logrus.Errorf("Could not find gtar in path: %e", err)
		return err
	}

	tarCmd := exec.Command(tarBin, "-C", altRootLocation, "-xzvf", tgzArchive)
	stdout, err := tarCmd.StdoutPipe()
	if err != nil {
		return err
	}

	// start the command after having set up the pipe
	if err := tarCmd.Start(); err != nil {
		return err
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	for in.Scan() {
		logrus.Printf(in.Text()) // write each line to your log, or anything you need
	}

	if err := in.Err(); err != nil {
		logrus.Printf("error: %s", err)
	}

	return tarCmd.Wait()
}

func installOSFromMediaFiles(saveLocation string) error {
	logrus.Infof("Installing OS from images under %s", saveLocation)
	var err error
	os.Mkdir(altMountLocation, os.ModeDir)
	rootImage := filepath.Join(saveLocation, solmediarootfileName)
	logrus.Infof("Mounting %s at %s", rootImage, altMountLocation)
	if err = mount.LoopDevice("ufs", altMountLocation, rootImage); err != nil {
		logrus.Errorf("Failure: %s", err)
		return err
	}
	logrus.Infof("Success")
	usrImage := filepath.Join(saveLocation, solusrfileName)
	usrMnt := filepath.Join(altMountLocation, "usr")
	logrus.Infof("Mounting %s at %s", usrImage, usrMnt)
	if err = mount.LoopDevice("hsfs", usrMnt, usrImage); err != nil {
		logrus.Errorf("Failure: %s", err)
		return err
	}
	logrus.Infof("Success")
	logrus.Infof("Copying everything recursively from %v", osfilelist)
	for _, dir := range osfilelist {
		if err := filepath.Walk(fmt.Sprintf("%s/%s", altMountLocation, dir), walkCopy); err != nil {
			return err
		}
	}
	return nil
}
