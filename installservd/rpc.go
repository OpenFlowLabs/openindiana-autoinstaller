package installservd

type InstallservdRPCReceiver struct {
	server *Installservd
}

func (r *InstallservdRPCReceiver) Ping(message string, reply *string) error {
	r.server.Echo.Logger.Print(message)
	*reply = "Pong"
	return nil
}

func hasAsset(name string) bool {
	_, ok := Assets[name]
	return ok
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

}
