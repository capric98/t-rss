package client

type ClientType struct {
	Name     string
	Client   interface{}
	Settings map[string]string
}
