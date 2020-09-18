package main

import (
	"os"
	"sync"

	"github.com/f5devcentral/bigip-tgw/as3"
	"github.com/f5devcentral/bigip-tgw/config"
	"github.com/f5devcentral/bigip-tgw/consul"
	"github.com/f5devcentral/bigip-tgw/gateway"
	slog "github.com/go-eden/slf4go"
)

func main() {

	var log = slog.NewLogger("init")
	slog.SetLevel(slog.DebugLevel)
	//load configuration
	c, err := config.Load()
	if err != nil {
		log.Errorf("unable to read configuration, error: %+v", err)
		os.Exit(0)
	}

	//Init as3manager
	agent := as3.CreateAgent()
	err = agent.Init(c.Bigip)
	if err != nil {
		log.Errorf("unable to init agent, error: %+v", err)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "remove" {
		err = agent.PostManager.DeletePartition([]string{"TGW_Tenant"})
		if err != nil {
			log.Errorf("unable to remove partition, error: %+v", err)
			os.Exit(0)
		}
		log.Info("removed AS3 partition TGW_Tenant")
		os.Exit(0)
	}

	//Init watcher
	watcher := consul.New()
	err = watcher.Init(c.Consul, c.Gateway.Name, c.Gateway.Namespace)
	if err != nil {
		log.Errorf("unable to create and configure Consul watcher, error: %+v", err)
		os.Exit(0)
	}

	//Init writer
	writer := gateway.New(c.Bigip, watcher.C, agent.ReqChan)
	//err = writer.Init(c.Bigip)
	//if err != nil {
	//	log.Errorf("unable to create and configure AS3 writer, error: %+v", err)
	//	os.Exit(0)
	//}
	defer writer.DeInit()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := watcher.Run()
		if err != nil {
			log.Panicf("error running consul watcher: %+v", err)
		}
	}()
	go func() {
		err := writer.Run()
		if err != nil {
			log.Panicf("error running F5 writer: %+v", err)
		}
	}()
	wg.Wait()
}
