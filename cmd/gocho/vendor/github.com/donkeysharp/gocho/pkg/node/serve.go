package node

import (
	"container/list"
	"github.com/donkeysharp/gocho/pkg/config"
	"time"
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
