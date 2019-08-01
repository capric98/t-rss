package client

type DeType struct {
	Name     string
	settings map[string]string
	port     int
}

func (c DeType) Init() error {
	return nil
}
