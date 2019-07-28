package client

type DeType struct {
	name string
}

func NewDeClient(m map[interface{}]interface{}) ClientType {
	var nc ClientType
	nc.Name = "Deluge"
	return nc
}
