// +build illumos

package installd

import (
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

func walkCopy(path string, info os.FileInfo, err error) error {
	dstpath := strings.Replace(path, altMountLocation, altRootLocation, -1)
	lsrcinfo, err := os.Lstat(path)
	if os.IsNotExist(err) {
		//Ignore Unexistent Directories
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		logrus.Tracef("Mkdir %s", dstpath)
		if err := os.Mkdir(dstpath, info.Mode()); err != nil {
			return err
		}
		srcStat := info.Sys().(*syscall.Stat_t)
		if err := syscall.Chmod(dstpath, srcStat.Mode); err != nil {
			return err
		}
		if err := syscall.Chown(dstpath, int(srcStat.Uid), int(srcStat.Gid)); err != nil {
			return err
		}
	} else if lsrcinfo.Mode()&os.ModeSymlink != 0 {
		//We have a Symlink thus Create it on the Target
		dstTarget, _ := os.Readlink(path)
		logrus.Tracef("Creating Symlink %s -> %s", dstpath, dstTarget)
		if err := os.Symlink(dstTarget, dstpath); err != nil {
			return err
		}
	} else {
		//We Have a regular File Copy it
		go copyFileExact(path, info, dstpath)
	}
	return nil
}

func copyFileExact(source string, srcInfo os.FileInfo, dest string) {
	logrus.Tracef("Copy %s -> %s", source, dest)
	src, err := os.Open(source)
	defer src.Close()
	if err != nil {
		logrus.Errorf("Cant open %s: %s", source, err)
		return
	}
	dst, err := os.Create(dest)
	defer dst.Close()
	if err != nil {
		logrus.Errorf("Cant open %s: %s", dest, err)
		return
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		logrus.Errorf("Can not copy %s -> %s: %s", source, dest, err)
	}
	//dst.Sync()
	srcStat := srcInfo.Sys().(*syscall.Stat_t)
	err = syscall.Chmod(dest, srcStat.Mode)
	err = syscall.Chown(dest, int(srcStat.Uid), int(srcStat.Gid))
	if err != nil {
		logrus.Errorf("Failed to set user/group/mode of %s: %s", dest, err)
	}
	//os.Chtimes(dest, time.Unix(int64(srcStat.Atim.Sec),int64(srcStat.Atim.Nsec)), time.Unix(int64(srcStat.Mtim.Sec),int64(srcStat.Mtim.Nsec)))
}
