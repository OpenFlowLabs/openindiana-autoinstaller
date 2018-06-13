package installservd

import (
	"io"
)

var Assets map[string]*Asset

type AssetType int

const (
	AssetTypeImage AssetType = iota
	AssetTypeConfig
	AssetTypeScript
	AssetTypeGeneric
	AssetTypeTemplate
)

const assetFileName = "assets.json"

type Asset struct {
	Path string
	Type AssetType
}

func DownloadAsset(source string, target io.Writer) error {

}

func InstallAsset(src io.Reader, target Asset) error {

}

func (i *Installservd) SaveAssetsToDisk() error {
	return saveToDisk(i.ServerHome, assetFileName, &Assets)
}

func (i *Installservd) LoadAssetsFromDisk() error {
	return loadFromDisk(i.ServerHome, assetFileName, &Assets, func() {
		Assets = make(map[string]*Asset)
	})
}
