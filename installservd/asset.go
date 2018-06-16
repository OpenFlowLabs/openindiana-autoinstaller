package installservd

import (
	"path/filepath"

	"github.com/satori/go.uuid"
)

var Assets map[string]*Asset

type AssetType int

const (
	AssetTypeImage AssetType = iota
	AssetTypeBinary
	AssetTypeTemplate
)

func getAssetTypeByName(name string) AssetType {
	switch name {
	case "image":
		return AssetTypeImage
	case "template":
		return AssetTypeTemplate
	default:
		return AssetTypeBinary
	}
}

const assetFileName = "assets.json"

type Asset struct {
	ID   uuid.UUID
	Path string
	Type AssetType
}

func (i *Installservd) getAssetPath(asset Asset) string {
	return filepath.Join(i.ServerHome, "assets", asset.Path)
}

func (i *Installservd) SaveAssetsToDisk() error {
	return saveToDisk(i.ServerHome, assetFileName, &Assets)
}

func (i *Installservd) LoadAssetsFromDisk() error {
	return loadFromDisk(i.ServerHome, assetFileName, &Assets, func() {
		Assets = make(map[string]*Asset)
	})
}
