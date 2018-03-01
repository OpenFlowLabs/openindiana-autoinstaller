// +build solaris

package installd

import (
	"fmt"

	"git.wegmueller.it/opencloud/opencloud/zfs"
	"git.wegmueller.it/opencloud/opencloud/zpool"
	"git.wegmueller.it/toasterson/glog"
	"github.com/satori/go.uuid"
)

func createAndMountZpool(conf *InstallConfiguration, noop bool) (err error) {
	for _, pool := range conf.Pools {
		if pool.Type == "" {
			pool.Type = "normal"
		}
		if noop {
			glog.Infof("Would Create Zpool %s as %s with options %v on disks %v", pool.Name, pool.Type, pool.Options, pool.Disks)
			continue
		}
		glog.Infof("Creating Pool %s as %s with option %v on %v", pool.Name, pool.Type, pool.Options, pool.Disks)
		_, err = zpool.CreatePool(pool.Name, pool.Options, true, pool.Type, pool.Disks, true)
		if err != nil {
			glog.Errf("Failure creating Pool %s: %s", pool.Name, err)
			return
		}
		glog.Infof("Success")
	}
	return
}

func createDatasets(conf *InstallConfiguration, noop bool) error {

	var err error
	if conf.InstallType != "bootenv" {
		if noop {
			glog.Infof("Would create Dataset %s/ROOT with mountpoint=legacy", conf.Rpool)
			glog.Infof("Would create Dataset %s/swap with blocksize=4k,size=%s", conf.Rpool, conf.SwapSize)
			glog.Infof("Would create Dataset %s/dump with size=%s", conf.Rpool, conf.DumpSize)
		} else {
			glog.Infof("Creating %s/ROOT with mountpoint=legacy", conf.Rpool)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/ROOT", conf.Rpool), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": "legacy"}, true); err != nil {
				glog.Errf("Failed: %s", err)
				return err
			}
			glog.Infof("Success")
			glog.Infof("Creating SWAP Space at %s/swap with blocksize=4k,size=%s", conf.Rpool, conf.SwapSize)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/swap", conf.Rpool), zfs.DatasetTypeVolume, map[string]string{"blocksize": "4k", "size": conf.SwapSize}, true); err != nil {
				glog.Errf("Failure: %s", err)
				return err
			}
			glog.Infof("Success")
			glog.Infof("Creating Dump device at %s/dump with size=%s", conf.Rpool, conf.DumpSize)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/dump", conf.Rpool), zfs.DatasetTypeVolume, map[string]string{"size": conf.DumpSize}, true); err != nil {
				glog.Errf("Failure: %s")
				return err
			}
			glog.Infof("Success")
		}
		for _, zfsDataset := range conf.Datasets {
			if noop {
				glog.Infof("Would Create Dataset %s with options %v", zfsDataset.Name, zfsDataset.Options)
				continue
			}
			glog.Infof("Creating %s with options %v", zfsDataset.Name, zfsDataset.Options)
			if _, err := zfs.CreateDataset(zfsDataset.Name, zfsDataset.Type, zfsDataset.Options, true); err != nil {
				glog.Errf("Failure %s", err)
				return err
			}
			glog.Infof("Success")
		}
	}
	if conf.InstallImage.Type != MediaTypeZImage {
		if noop {
			glog.Infof("Would create Boot Environment %s/ROOT/%s", conf.Rpool, conf.BEName)
		} else {
			glog.Infof("Creating BootEnvironment %s/ROOT/%s", conf.Rpool, conf.BEName)
			var bootenv *zfs.Dataset
			if bootenv, err = zfs.CreateDataset(conf.GetRootDataSetName(), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": altRootLocation}, false); err == nil {
				u1, _ := uuid.NewV4()
				bootenv.SetProperty("org.opensolaris.libbe:uuid", u1.String())
				glog.Infof("Success")
			} else {
				glog.Errf("Failure: %s", err)
				return err
			}
		}
	} else {
		err := fmt.Errorf("media type %s not supported yet", MediaTypeZImage)
		glog.Errf("%s", err)
		return err
	}
	if noop {
		glog.Infof("Would set bootfs to %s/ROOT/%s", conf.Rpool, conf.BEName)
	} else {
		glog.Infof("Setting bootfs to %s/ROOT/%s", conf.Rpool, conf.BEName)
		rpool := zpool.OpenPool(conf.Rpool)
		if err = rpool.SetProperty("bootfs", fmt.Sprintf("%s/ROOT/%s", conf.Rpool, conf.BEName)); err != nil {
			glog.Errf("Failure: %s", err)
			return err
		}
		glog.Infof("Success")
	}
	return nil
}
