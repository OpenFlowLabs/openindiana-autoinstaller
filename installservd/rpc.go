package installservd

type InstallservdRPCReceiver struct {
	server *Installservd
}

func (r *InstallservdRPCReceiver) Ping(message string, reply *string) error {
	r.server.Echo.Logger.Print(message)
	*reply = "Pong"
	return nil
}
