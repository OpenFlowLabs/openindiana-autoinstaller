package installservd

import (
	"os"

	"net"

	"net/rpc"
	"strings"

	"net/http"
	"path/filepath"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

var (
	serverDirectories = []string{"config", "assets", "tmp"}
)

type Installservd struct {
	Echo        *echo.Echo
	ServerHome  string
	Socket      net.Listener
	SocketPath  string
	RPCReceiver *RPCReceiver
	runRPC      bool
}

func New() (*Installservd, error) {
	srvHome := os.ExpandEnv(viper.GetString("home"))
	i := &Installservd{
		Echo:       echo.New(),
		ServerHome: srvHome,
	}
	if err := i.LoadProfilesFromDisk(); err != nil {
		return nil, err
	}
	if err := i.LoadAssetsFromDisk(); err != nil {
		return nil, err
	}
	if err := i.setupWebServer(); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Installservd) setupWebServer() error {
	if _, err := os.Stat(i.ServerHome); os.IsNotExist(err) {
		os.Mkdir(i.ServerHome, 0755)
	}
	if err := os.Chdir(i.ServerHome); err != nil {
		return err
	}
	for _, dir := range serverDirectories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.Mkdir(dir, 0755)
		}
	}
	t := &Template{}
	t.Load(i)
	i.Echo.Renderer = t
	fs := http.FileServer(http.Dir(filepath.Join(i.ServerHome, "assets")))
	i.Echo.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", fs)))
	i.Echo.GET("/profiles", listProfiles)
	i.Echo.GET("/profiles/:name", getProfile)
	i.Echo.GET("/profiles/:name/config.json", getProfileConfig)
	i.Echo.GET("/profiles/:name/:template", getTemplate(i))

	return nil
}

func (i *Installservd) StartRPC(sock string) (err error) {
	i.SocketPath = sock
	i.RPCReceiver = &RPCReceiver{server: i}
	i.runRPC = true
	if _, err := os.Stat(i.SocketPath); !os.IsNotExist(err) {
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
	i.Socket = socket
	rpcSrv := rpc.NewServer()
	rpcSrv.Register(i.RPCReceiver)
	go func() {
		for i.runRPC {
			conn, err := i.Socket.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					i.runRPC = false
				} else {
					i.Echo.Logger.Printf("rpc.Serve: accept: %s", err.Error())
				}
				return
			}
			go rpcSrv.ServeConn(conn)
		}
	}()
	return nil
}

func (i *Installservd) StopRPC() error {
	i.runRPC = false
	err := i.Socket.Close()
	if err != nil {
		os.Remove(i.SocketPath)
		return err
	}
	return os.Remove(i.SocketPath)
}
