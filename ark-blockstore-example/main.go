package main

import (
	"github.com/heptio/ark/pkg/plugin"
)

func main() {
	plugin.Serve(plugin.NewBlockStorePlugin(NewNoOpBlockStore()))
}
