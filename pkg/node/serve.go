package node

import (
	"container/list"
	"time"

	"github.com/temorfeouz/gocho/pkg/config"
)

func startAnnouncer(conf *config.Config, nodeList *list.List) {
	announcer := &Announcer{
		config: conf,
	}
	announcer.Start(nodeList)
}

func Serve(conf *config.Config) {
	nodeList := list.New()

	go startAnnouncer(conf, nodeList)
	go fileServe(conf)
	go dashboardServe(conf, nodeList)

	for {
		time.Sleep(time.Minute * 15)
	}
}
