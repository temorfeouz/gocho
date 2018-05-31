package main

//go:generate go-bindata -o ../../assets/assets_gen.go -pkg assets ../../ui/build/...

import (
	"os"

	"github.com/temorfeouz/gocho/pkg/cmds"
)

func main() {
	cmds.New().Run(os.Args)
}
