package installservd

type Group struct {
	Name      string
	Profile   Profile
	Selectors map[string]string
	Metadata  map[string]string
}
