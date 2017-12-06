// +build solaris

package installd

import (
	"fmt"

	"git.wegmueller.it/opencloud/opencloud/zfs"
	"git.wegmueller.it/opencloud/opencloud/zpool"
	"github.com/satori/go.uuid"
	"github.com/toasterson/mozaik/util"
	"git.wegmueller.it/toasterson/glog"
)

func createAndMountZpool(conf *InstallConfiguration, noop bool) (err error) {
	if noop {
		glog.Infof("Would Create Zpool %s with args %v with disks %v", conf.RPoolName, conf.PoolArgs, conf.PoolType)
		return
	}
	_, err = zpool.CreatePool(conf.RPoolName, conf.PoolArgs, true, conf.PoolType, conf.Disks, true)
	if err != nil {
		return
	}
	return
}

func createDatasets(conf *InstallConfiguration, noop bool) error {
	if conf.SwapSize == "" {
		conf.SwapSize = "2g"
	}
	if conf.DumpSize == "" {
		conf.DumpSize = conf.SwapSize
	}
	if conf.BEName == "" {
		conf.BEName = "openindiana"
	}
	var err error
	if conf.InstallType != "bootenv" {
		if noop {
			glog.Infof("Would create Dataset %s/ROOT with mountpoint=legacy", conf.RPoolName)
			glog.Infof("Would create Dataset %s/swap with blocksize=4k,size=%s", conf.RPoolName, conf.SwapSize)
			glog.Infof("Would create Dataset %s/dump with size=%s", conf.RPoolName, conf.DumpSize)
		} else {
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/ROOT", conf.RPoolName), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": "legacy"}, true); err != nil {
				return err
			}

			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/swap", conf.RPoolName), zfs.DatasetTypeVolume, map[string]string{"blocksize": "4k", "size": conf.SwapSize}, true); err != nil {
				return err
			}

			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/dump", conf.RPoolName), zfs.DatasetTypeVolume, map[string]string{"size": conf.DumpSize}, true); err != nil {
				return err
			}
		}

		//TODO Zfs Layout Creation
	}
	if conf.MediaType != MediaTypeZImage {
		if noop {
			glog.Infof("Would create Boot Environment %s/ROOT/%s", conf.RPoolName, conf.BEName)
		} else {
			var bootenv *zfs.Dataset
			if bootenv, err = zfs.CreateDataset(fmt.Sprintf("%s/ROOT/%s", conf.RPoolName, conf.BEName), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": altRootLocation}, true); err == nil {
				u1 := uuid.NewV4()
				bootenv.SetProperty("org.opensolaris.libbe:uuid", u1.String())
			} else {
				return err
			}
		}
	}
	if noop {
		glog.Infof("Would set bootfs to %s/ROOT/%s", conf.RPoolName, conf.BEName)
	} else {
		rpool := zpool.OpenPool(conf.RPoolName)
		if err = rpool.SetProperty("bootfs", fmt.Sprintf("%s/ROOT/%s", conf.RPoolName, conf.BEName)); err != nil {
			return err
		}
	}
	return nil
}
