package installservd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"git.wegmueller.it/opencloud/installer/fileutils"
)

func saveToDisk(home, fileName string, obj interface{}) error {
	var fileObj *os.File
	var err error
	path := filepath.Join(home, "config", fileName)
	pathBackup := path + ".bak"
	if _, err = fileutils.CopyFile(path, pathBackup); err != nil {
		return err
	}
	if fileObj, err = os.Create(path); err != nil {
		return err
	}
	defer func() {
		// Best Effort we Ignore write error here since in that case there is something very wrong
		fileObj.Close()
		if err == nil {
			os.Remove(pathBackup)
		} else {
			fileutils.CopyFile(pathBackup, path)
		}
	}()
	err = json.NewEncoder(fileObj).Encode(obj)
	return err
}

func loadFromDisk(home, fileName string, obj interface{}, init func()) error {
	var file *os.File
	var err error
	path := filepath.Join(home, "config", fileName)
	if file, err = os.Open(path); err != nil {
		if os.IsNotExist(err) {
			// We have no Config for that. This can be expected
			// Just initialize as empty
			init()
			return nil
		} else {
			return err
		}
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(obj)
}
