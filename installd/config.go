package installd

import (
	"encoding/json"
	"fmt"

	"git.wegmueller.it/opencloud/opencloud/zfs"
	"github.com/OpenFlowLabs/openindiana-autoinstaller/devprop"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	MediaTypeSolNetBoot = "solnetboot"
	MediaTypeSolCDrom   = "solcdrom"
	MediaTypeZImage     = "zimage"
	MediaTypeTGZ        = "tgz"
)

const (
	InstallTypeBootEnv     = "bootenv"
	InstallTypeEFIFullDisk = "fulldisk"
)

type InstallImage struct {
	Type string `json:"type"` //Valid Values are SolNetboot, SolCdrom, SolUSB, ZImage, TGZ(Default)
	URL  string `json:"url"`  //Where to get the Image from
}

type ZFSLayout struct {
	Name    string          `json:"name"`    //Name of the Dataset
	Type    zfs.DatasetType `json:"type"`    //Type of the Dataset Available are Filesystem(Default), Volume
	Options zfs.Properties  `json:"options"` //Options to create the Dataset with
}

func (layout *ZFSLayout) UnmarshalJSON(input []byte) error {
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
	InstallType  string          `json:"install_type"`  //Possible options are efi, bootenv, fulldisk
	Pools        []ZPool         `json:"pools"`         //What pools to create on which disks
	InstallImage InstallImage    `json:"install_image"` //See InstallImage struct
	Datasets     []ZFSLayout     `json:"datasets"`      //The Partition Layout for ZFS. e.g Where is /var /etc and others located
	Rpool        string          `json:"rpool"`         //Name of the root pool By Default(rpool)
	BEName       string          `json:"be_name"`       //Name of the new Boot Environment defaults to openindiana
	SwapSize     string          `json:"swap_size"`     //Size of the SWAP Partition defaults to 2g
	DumpSize     string          `json:"dump_size"`     //Size of the Dump Partition defaults to swap_size
	BootLoader   string          `json:"boot_loader"`   //Valid values are Loader and Grub
	Net          NetworkSettings `json:"net"`           //The Network Interfaces of the Box
	TimeZone     string          `json:"time_zone"`     //Timezone to setup
	Locale       string          `json:"locale"`        //Locale like en_US.UTF-8 or de_CH.UTF-8
	RootPWClear  string          `json:"root_pw_clear"` //The clear string root password
	RootPW       string          `json:"root_pw"`       //The hashed
	Hostname     string          `json:"hostname"`      //The Hostname of the machine
	Keymap       string          `json:"keymap"`        //Keymap the Machine will have e.g. Swiss-German
}

func (conf *InstallConfiguration) GetRootDataSetName() string {
	return fmt.Sprintf("%s/ROOT/%s", conf.Rpool, conf.BEName)
}

func (conf *InstallConfiguration) FillUnSetValues() {
	if conf.SwapSize == "" {
		conf.SwapSize = "2g"
	}
	if conf.DumpSize == "" {
		conf.DumpSize = conf.SwapSize
	}
	if conf.BEName == "" {
		conf.BEName = "openindiana"
	}
	if conf.Rpool == "" {
		conf.Rpool = "rpool"
	}
	//Assume that we want the Media URL from devprop if it is not in the config
	if conf.InstallImage.URL == "" {
		conf.InstallImage.URL = devprop.GetValue("install_media")
	}

	if conf.RootPW == "" {
		if conf.RootPWClear != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(conf.RootPWClear), 5)
			if err != nil {
				logrus.Errorf("Could not hash password: This should not happen. Will terminate now. %s", err)
				panic(err)
			}
			conf.RootPW = string(hash)
		}
	}
}
