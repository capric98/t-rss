package receiver

// Receiver interface
type Receiver interface {
	Push([]byte, interface{}) error
}
