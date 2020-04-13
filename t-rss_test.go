package trss

import "testing"

func TestWithConfigFile(t *testing.T) {
	WithConfigFile("config.example.yml", "trace", true)
}
