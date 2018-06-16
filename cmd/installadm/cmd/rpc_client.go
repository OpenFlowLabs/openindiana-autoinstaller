package cmd

import (
	"net/rpc"
)

func rpcDialServer(socket, function string, args, reply interface{}) error {
	client, err := rpc.Dial("unix", socket)
	if err != nil {
		return err
	}
	if err := client.Call(function, args, reply); err != nil {
		return err
	}
	return nil
}
