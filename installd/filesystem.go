// +build illumos

package installd

import (
	"fmt"

	"git.wegmueller.it/opencloud/opencloud/zfs"
	"git.wegmueller.it/opencloud/opencloud/zpool"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func createAndMountZpool(conf *InstallConfiguration, noop bool) (err error) {
	for _, pool := range conf.Pools {
		if pool.Type == "" {
			pool.Type = "normal"
		}
		if noop {
			logrus.Infof("Would Create Zpool %s as %s with options %v on disks %v", pool.Name, pool.Type, pool.Options, pool.Disks)
			continue
		}
		logrus.Infof("Creating Pool %s as %s with option %v on %v", pool.Name, pool.Type, pool.Options, pool.Disks)
		_, err = zpool.CreatePool(pool.Name, pool.Options, true, pool.Type, pool.Disks, true)
		if err != nil {
			logrus.Errorf("Failure creating Pool %s: %s", pool.Name, err)
			return
		}
		logrus.Infof("Success")
	}
	return
}

func createDatasets(conf *InstallConfiguration, noop bool) error {

	var err error
	if conf.InstallType != InstallTypeBootEnv {
		if noop {
			logrus.Infof("Would create Dataset %s/ROOT with mountpoint=legacy", conf.Rpool)
			logrus.Infof("Would create Dataset %s/swap with blocksize=4k,size=%s", conf.Rpool, conf.SwapSize)
			logrus.Infof("Would create Dataset %s/dump with size=%s", conf.Rpool, conf.DumpSize)
		} else {
			logrus.Infof("Creating %s/ROOT with mountpoint=legacy", conf.Rpool)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/ROOT", conf.Rpool), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": "legacy"}); err != nil {
				logrus.Errorf("Failed: %s", err)
				return err
			}
			logrus.Infof("Success")
			logrus.Infof("Creating SWAP Space at %s/swap with blocksize=4k,size=%s", conf.Rpool, conf.SwapSize)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/swap", conf.Rpool), zfs.DatasetTypeVolume, map[string]string{"blocksize": "4k", "size": conf.SwapSize}); err != nil {
				logrus.Errorf("Failure: %s", err)
				return err
			}
			logrus.Infof("Success")
			logrus.Infof("Creating Dump device at %s/dump with size=%s", conf.Rpool, conf.DumpSize)
			if _, err = zfs.CreateDataset(fmt.Sprintf("%s/dump", conf.Rpool), zfs.DatasetTypeVolume, map[string]string{"size": conf.DumpSize}); err != nil {
				logrus.Errorf("Failure: %s")
				return err
			}
			logrus.Infof("Success")
		}
		for _, zfsDataset := range conf.Datasets {
			if noop {
				logrus.Infof("Would Create Dataset %s with options %v", zfsDataset.Name, zfsDataset.Options)
				continue
			}
			logrus.Infof("Creating %s with options %v", zfsDataset.Name, zfsDataset.Options)
			if _, err := zfs.CreateDataset(zfsDataset.Name, zfsDataset.Type, zfsDataset.Options); err != nil {
				logrus.Errorf("Failure %s", err)
				return err
			}
			logrus.Infof("Success")
		}
	}
	if conf.InstallImage.Type != MediaTypeZImage {
		if noop {
			logrus.Infof("Would create Boot Environment %s/ROOT/%s", conf.Rpool, conf.BEName)
		} else {
			logrus.Infof("Creating BootEnvironment %s/ROOT/%s", conf.Rpool, conf.BEName)
			var bootenv *zfs.Dataset
			if bootenv, err = zfs.CreateDataset(conf.GetRootDataSetName(), zfs.DatasetTypeFilesystem, map[string]string{"mountpoint": altRootLocation}); err == nil {
				u1, err := uuid.NewV4()
				if err != nil {
					return err
				}

				if err := bootenv.SetProperty("org.opensolaris.libbe:uuid", u1.String()); err != nil {
					return err
				}

				logrus.Infof("Success")
			} else {
				logrus.Errorf("Failure: %s", err)
				return err
			}
		}
	} else {
		err := fmt.Errorf("media type %s not supported yet", MediaTypeZImage)
		logrus.Errorf("%s", err)
		return err
	}
	if noop {
		logrus.Infof("Would set bootfs to %s/ROOT/%s", conf.Rpool, conf.BEName)
	} else {
		logrus.Infof("Setting bootfs to %s/ROOT/%s", conf.Rpool, conf.BEName)
		rpool := zpool.OpenPool(conf.Rpool)
		if err = rpool.SetProperty("bootfs", fmt.Sprintf("%s/ROOT/%s", conf.Rpool, conf.BEName)); err != nil {
			logrus.Errorf("Failure: %s", err)
			return err
		}
		logrus.Infof("Success")
	}
	return nil
}
