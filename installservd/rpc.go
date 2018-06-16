package installservd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (r *InstallservdRPCReceiver) AddAsset(args AddAssetArg, reply *string) error {
	var err error
	if args.Source == "" && args.Content == nil {
		err = fmt.Errorf("either Source or Content needs to be present")
		*reply = err.Error()
		return err
	}

	if args.Path == "" && args.Name == "" {
		err = fmt.Errorf("either Name or Path need to be defined or the Asset can not be saved")
		*reply = err.Error()
		return err
	}

	if args.Name == "" {
		split := strings.Split(args.Path, "/")
		args.Name = split[len(split)-1]
	}

	tmpFileName := filepath.Join(r.server.ServerHome, "tmp", args.Name)
	var tmpFile *os.File

	if strings.Contains(args.Source, "http") {
		//Download Asset from HTTP and Update Source to point to local file
		if err = fileutils.HTTPDownload(args.Source, tmpFileName); err != nil {
			*reply = err.Error()
			return err
		}
		if tmpFile, err = os.Open(tmpFileName); err != nil {
			*reply = err.Error()
			return err
		}
	} else {
		if tmpFile, err = os.Create(tmpFileName); err != nil {
			*reply = err.Error()
			return err
		}
		if _, err = io.Copy(bytes.NewBuffer(args.Content), tmpFile); err != nil {
			*reply = err.Error()
			return err
		}
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFileName)
	}()

	assetUUID, _ := uuid.NewV4()
	asset := Asset{
		ID:   assetUUID,
		Path: args.Path,
		Type: getAssetTypeByName(args.Type),
	}

	var finalFile *os.File
	if finalFile, err = os.Create(r.server.getAssetPath(asset)); err != nil {
		*reply = err.Error()
		return err
	}

	if _, err = io.Copy(tmpFile, finalFile); err != nil {
		*reply = err.Error()
		return err
	}

	Assets[args.Path] = &asset
	if err := r.server.SaveAssetsToDisk(); err != nil {
		*reply = err.Error()
		return err
	}

	*reply = "success"
	return nil
}

func (r *InstallservdRPCReceiver) ListProfiles(args string, reply *[]Profile) error {
	*reply = Profiles
	return nil
}
