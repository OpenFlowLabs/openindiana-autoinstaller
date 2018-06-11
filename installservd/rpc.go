package installservd

import (
	"net"
	"net/rpc"
	"os"
	"strings"
)

func (d *Installservd) HandleRPC(sock string) (err error) {
	if _, err := os.Stat(d.SocketPath); !os.IsNotExist(err) {
		//Best effort
		err = os.Remove(sock)
		if err != nil {
			return err
		}
	}
	socket, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	d.Socket = socket
	rpcSrv := rpc.NewServer()
	rpcSrv.Register(d)
	go func() {
		for d.runRPC {
			conn, err := d.Socket.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					d.runRPC = false
				} else {
					d.Echo.Logger.Printf("rpc.Serve: accept: %s", err.Error())
				}
				return
			}
			go rpcSrv.ServeConn(conn)
		}
	}()
	return nil
}

func (d *Installservd) Ping(message string, reply *string) error {
	d.Echo.Logger.Print(message)
	*reply = "Pong"
	return nil
}

func (d *Installservd) StopRPC() error {
	d.runRPC = false
	err := d.Socket.Close()
	if err != nil {
		os.Remove(d.SocketPath)
		return err
	}
	return os.Remove(d.SocketPath)
}
