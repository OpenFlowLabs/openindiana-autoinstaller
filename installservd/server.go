package installservd

import (
	"os"

	"net"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

var (
	serverDirectories = []string{"config", "files", "tmp"}
)

type Installservd struct {
	Echo       *echo.Echo
	ServerHome string
	Socket     net.Listener
	SocketPath string
	runRPC     bool
}

func New() (*Installservd, error) {
	srvHome := os.ExpandEnv(viper.GetString("home"))
	i := &Installservd{
		Echo:       echo.New(),
		ServerHome: srvHome,
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
	i.Echo.Static("/files", "files")
	return nil
}
