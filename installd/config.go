package installd

import (
	"encoding/json"
	"fmt"

	"git.wegmueller.it/opencloud/installer/net"
	"git.wegmueller.it/opencloud/opencloud/zfs"
)

const (
	MediaTypeSolNetBoot = "solnetboot"
	MediaTypeSolCDrom   = "solcdrom"
	MediaTypeSolUSB     = "solusb"
	MediaTypeZImage     = "zimage"
	MediaTypeACI        = "ACI"
)

const (
	InstallTypeBootEnv     = "bootenv"
	InstallTypeEFIFullDisk = "fulldisk"
)

type InstallImage struct {
	Type string `json:"type"` //Valid Values are SolNetboot, SolCdrom, SolUSB, ZImage, ACI(Default)
	URL  string `json:"url"`  //Where to get the Image from
}

type ZFSLayout struct {
	Name    string          `json:"name"`    //Name of the Dataset
	Type    zfs.DatasetType `json:"type"`    //Type of the Dataset Available are Filesystem(Default), Volume
	Options zfs.Properties  `json:"options"` //Options to create the Dataset with
}

func (layout ZFSLayout) UnmarshalJSON(input []byte) error {
	var zfsLayoutJson map[string]*json.RawMessage
	if err := json.Unmarshal(input, &zfsLayoutJson); err != nil {
		return err
	}
	for key, val := range zfsLayoutJson {
		switch key {
		case "name":
			if err := json.Unmarshal(*val, &layout.Name); err != nil {
				return err
			}
		case "type":
			var stringFSType string
			if err := json.Unmarshal(*val, &stringFSType); err != nil {
				return err
			}
			if stringFSType == "volume" {
				layout.Type = zfs.DatasetTypeVolume
			} else {
				layout.Type = zfs.DatasetTypeFilesystem
			}
		case "options":
			if err := json.Unmarshal(*val, &layout.Options); err != nil {
				return err
			}
		default:
			return fmt.Errorf("could not deserialize zfs layout %s is an unrecognized key", key)
		}
	}
	return nil
}

type ZPool struct {
	Name    string         `json:"name"`    //Name of the pool
	Disks   []string       `json:"disks"`   //The disks to use to back the Zpool
	Options zfs.Properties `json:"options"` //The ZFS Options to create the pool with
	Type    string         `json:"type"`    //Type of the pool e.g. mirror, raidz raidz2. Leave empty for normal
}

type InstallConfiguration struct {
	InstallType  string              `json:"install_type"`  //Possible options are efi, bootenv, fulldisk
	Pools        []ZPool             `json:"pools"`         //What pools to create on which disks
	InstallImage InstallImage        `json:"install_image"` //See InstallImage struct
	Datasets     []ZFSLayout         `json:"datasets"`      //The Partition Layout for ZFS. e.g Where is /var /etc and others located
	Rpool        string              `json:"rpool"`         //Name of the root pool By Default(rpool)
	BEName       string              `json:"be_name"`       //Name of the new Boot Environment defaults to openindiana
	SwapSize     string              `json:"swap_size"`     //Size of the SWAP Partition defaults to 2g
	DumpSize     string              `json:"dump_size"`     //Size of the Dump Partition defaults to swap_size
	BootLoader   string              `json:"boot_loader"`   //Valid values are Loader and Grub
	Net          net.NetworkSettings `json:"net"`           //The Network Interfaces of the Box
}
