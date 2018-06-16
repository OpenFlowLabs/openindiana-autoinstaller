package installservd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"io/ioutil"

	"git.wegmueller.it/opencloud/installer/fileutils"
	"github.com/satori/go.uuid"
)

type InstallservdRPCReceiver struct {
	server *Installservd
}

func (r *InstallservdRPCReceiver) Ping(message string, reply *string) error {
	r.server.Echo.Logger.Print(message)
	*reply = "Pong"
	return nil
}

func hasAssets(names ...string) bool {
	ok := true
	for _, name := range names {
		_, ok = Assets[name]
		if !ok {
			return false
		}
	}
	return ok
}

func getAssets(names ...string) []*Asset {
	assets := make([]*Asset, 0)
	for _, name := range names {
		assets = append(assets, Assets[name])
	}
	return assets
}

func (r *InstallservdRPCReceiver) AddProfile(profile Profile, reply *string) error {
	if err := r.server.AddProfile(profile); err != nil {
		*reply = err.Error()
		return err
	}
	if err := r.server.SaveProfilesToDisk(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = "success"
	return nil
}

type AddAssetArg struct {
	Source  string
	Content []byte
	Path    string
	Name    string
	Type    string
}

func (r *InstallservdRPCReceiver) AddAsset(args AddAssetArg, reply *string) (err error) {

	defer func() {
		if err != nil {
			*reply = err.Error()
		} else {
			*reply = "Operation suceeded"
		}
	}()

	if args.Source == "" && args.Content == nil {
		return fmt.Errorf("either Source or Content needs to be present")
	}

	if args.Path == "" && args.Name == "" {
		return fmt.Errorf("either Name or Path need to be defined or the Asset can not be saved")
	}

	if args.Name == "" {
		split := strings.Split(args.Path, "/")
		args.Name = split[len(split)-1]
	}

	//Sanity Check No two assets with the same name should exist.
	if _, ok := Assets[args.Path]; ok {
		return fmt.Errorf("asset %s already exists", args.Path)
	}

	tmpFileName := filepath.Join(r.server.ServerHome, "tmp", args.Name)

	if strings.Contains(args.Source, "http") {
		//Download Asset from HTTP and Update Source to point to local file
		if err = fileutils.HTTPDownload(args.Source, tmpFileName); err != nil {
			return
		}
	} else {
		if err = ioutil.WriteFile(tmpFileName, args.Content, 0644); err != nil {
			return
		}
	}

	defer func() {
		os.Remove(tmpFileName)
	}()

	assetUUID, _ := uuid.NewV4()
	asset := Asset{
		ID:   assetUUID,
		Path: args.Path,
		Type: getAssetTypeByName(args.Type),
	}

	path := r.server.getAssetPath(asset)
	dirPath, _ := filepath.Split(path)
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return
	}

	if _, err = fileutils.CopyFile(tmpFileName, path); err != nil {
		return
	}

	Assets[args.Path] = &asset
	if err := r.server.SaveAssetsToDisk(); err != nil {
		return err
	}
	return nil
}

func (r *InstallservdRPCReceiver) ListProfiles(args string, reply *[]Profile) error {
	*reply = Profiles
	return nil
}
