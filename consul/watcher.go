package consul

import (
	"crypto/x509"
	"sync"
	"time"

	slog "github.com/go-eden/slf4go"

	"github.com/hashicorp/consul/api"
)

const (
	errorWaitTime = 5 * time.Second
)

var log = slog.NewLogger("consul-watcher")

type ConsulConfig struct {
	// Address is the address of the Consul server
	Address string

	// Scheme is the URI scheme for the Consul server
	Scheme string

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string

	// Transport is the Transport to use for the http client.
	//Transport *http.Transport

	// HttpClient is the client to use. Default will be
	// used if not provided.
	//HttpClient *http.Client

	// HttpAuth is the auth info to use for http access.
	//HttpAuth *HttpBasicAuth

	// WaitTime limits how long a Watch will block. If not provided,
	// the agent default values will be used.
	//WaitTime time.Duration

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string

	// TokenFile is a file containing the current token to use for this client.
	// If provided it is read once at startup and never again.
	//TokenFile string

	// Namespace is the name of the namespace to send along for the request
	// when no other Namespace ispresent in the QueryOptions
	Namespace string

	//TLSConfig TLSConfig
}

type service struct {
	name           string
	instances      []*api.ServiceEntry
	intentions     []*api.Intention
	gatewayService *api.GatewayService
	leaf           *certLeaf

	ready sync.WaitGroup
	done  bool
}

type certLeaf struct {
	Cert []byte
	Key  []byte

	done bool
}

//Watcher struct for TG config
type Watcher struct {
	settings api.Config

	id        string
	name      string
	namespace string
	address   string
	port      int
	consul    *api.Client
	token     string
	C         chan Config

	lock  sync.Mutex
	ready sync.WaitGroup

	services   map[string]*service
	certCAs    [][]byte
	certCAPool *x509.CertPool
	leaf       *certLeaf

	update chan struct{}
}

//New Watcher
func New() *Watcher {

	log.Info("creating new Consul watcher")
	return &Watcher{
		C:        make(chan Config),
		services: make(map[string]*service),
		update:   make(chan struct{}, 1),
	}
}

func (w *Watcher) Init(c ConsulConfig, gatewayName string, namespace string) error {
	var err error

	log.Infof("initializing Consul watcher for gateway: %+v", gatewayName)
	w.name = gatewayName
	w.namespace = namespace
	w.settings = *api.DefaultConfig()
	w.settings.Address = c.Address
	w.settings.Scheme = c.Scheme
	w.settings.Token = c.Token
	w.settings.Namespace = c.Namespace
	w.consul, err = api.NewClient(&w.settings)
	if err != nil {
		return err
	}
	return nil
}

//Run Watcher
func (w *Watcher) Run() error {

	//Debug
	log.Debugf("running watcher for gateway: %s\n", w.name)

	w.ready.Add(3)

	go w.watchService(w.name, true, "terminating-gateway")
	go w.watchGateway()
	go w.watchCA()

	w.ready.Wait()

	for range w.update {
		w.C <- w.genCfg()
	}

	return nil
}

//Reload Configuration
func (w *Watcher) Reload() {
	w.C <- w.genCfg()
}

func (w *Watcher) watchLeaf(service string, first bool) {
	log.Debugf("watching leaf cert for %s", service)
	dFirst := true
	var lastIndex uint64
	for {
		if w.services[service] == nil {
			return
		} else if w.services[service].done {
			return
		}
		cert, meta, err := w.consul.Agent().ConnectCALeaf(service, &api.QueryOptions{
			WaitTime:  10 * time.Minute,
			WaitIndex: lastIndex,
		})
		if err != nil {
			log.Errorf("consul error fetching leaf cert for service %s: %s", service, err)
			time.Sleep(errorWaitTime)
			if meta != nil {
				if meta.LastIndex < lastIndex || meta.LastIndex < 1 {
					lastIndex = 0
				}
			}
			continue
		}

		changed := lastIndex != meta.LastIndex
		lastIndex = meta.LastIndex

		if changed {
			log.Infof("leaf cert for service %s changed, serial: %s, valid before: %s, valid after: %s", service, cert.SerialNumber, cert.ValidBefore, cert.ValidAfter)
			w.lock.Lock()
			if w.services[service].leaf == nil {
				w.services[service].leaf = &certLeaf{}
			}
			w.services[service].leaf.Cert = []byte(cert.CertPEM)
			w.services[service].leaf.Key = []byte(cert.PrivateKeyPEM)
			w.lock.Unlock()
			if dFirst {
				w.services[service].ready.Done()
				dFirst = false
			} else {
				w.notifyChanged()
			}
		}

		if first {
			log.Infof("leaf cert for %s ready", service)
			w.ready.Done()
			first = false
		}
	}
}

func (w *Watcher) watchIntention(service string, first bool) {
	log.Debugf("watching intentions for %s", service)
	dFirst := true
	var lastIndex uint64

	for {
		if w.services[service] == nil {
			return
		} else if w.services[service].done {
			return
		}
		intentionList, meta, err := w.consul.Connect().Intentions(&api.QueryOptions{
			WaitTime:  10 * time.Minute,
			WaitIndex: lastIndex,
			Filter:    "DestinationName==" + service,
		})
		if err != nil {
			log.Errorf("consul error fetching intentions for service %s: %s", service, err)
			time.Sleep(errorWaitTime)
			if meta != nil {
				if meta.LastIndex < lastIndex || meta.LastIndex < 1 {
					lastIndex = 0
				}
			}
			continue
		}

		changed := lastIndex != meta.LastIndex
		lastIndex = meta.LastIndex

		if changed {
			log.Infof("intentions for service %s changed", service)
			w.lock.Lock()
			w.services[service].intentions = intentionList
			w.lock.Unlock()
			if dFirst {
				w.services[service].ready.Done()
				dFirst = false
			} else {
				w.notifyChanged()
			}
		}

		if first {
			log.Infof("intentions for %s ready", service)
			w.ready.Done()
			first = false
		}
	}
}

func (w *Watcher) watchGateway() {
	var lastIndex uint64
	first := true
	for {
		gwServices, meta, err := w.consul.Catalog().GatewayServices(w.name, &api.QueryOptions{
			WaitTime:  10 * time.Minute,
			WaitIndex: lastIndex,
		})
		if err != nil {
			log.Errorf("error fetching linked services for gateway %s: %s", w.name, err)
			time.Sleep(errorWaitTime)
			if meta != nil {
				if meta.LastIndex < lastIndex || meta.LastIndex < 1 {
					lastIndex = 0
				}
			}
			continue
		}

		changed := lastIndex != meta.LastIndex
		lastIndex = meta.LastIndex

		if changed {
			log.Infof("linked services changed for gateway %s", w.name)
			if first && len(gwServices) == 0 {
				log.Infof("no linked services defined for gateway %s", w.name)
				continue
			}
			w.handleProxyChange(first, &gwServices)
		}
		if first {
			log.Infof("linked services for %s ready", w.name)
			first = false
			w.ready.Done()
		}
	}
}

func (w *Watcher) watchService(service string, first bool, kind string) {
	log.Infof("watching downstream: %s", service)
	dFirst := true
	var lastIndex uint64
	var nSpace string
	for {
		if kind != "terminating-gateway" {
			if &w.services[service].gatewayService.Service.Namespace != nil {
				nSpace = w.services[service].gatewayService.Service.Namespace
			}
			if w.services[service] == nil {
				return
			} else if w.services[service].done {
				return
			}
		} else {
			nSpace = w.namespace
		}

		srv, meta, err := w.consul.Health().Service(service, "", false, &api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  10 * time.Minute,
			Namespace: nSpace,
		})
		if err != nil {
			log.Errorf("error fetching service %s definition: %s", service, err)
			time.Sleep(errorWaitTime)
			if meta != nil {
				if meta.LastIndex < lastIndex || meta.LastIndex < 1 {
					lastIndex = 0
				}
			}
			continue
		}

		changed := lastIndex != meta.LastIndex
		lastIndex = meta.LastIndex

		if changed {
			log.Debugf("service %s changed", service)
			if len(srv) == 0 {
				log.Infof("no service definition found for: %s", service)
				continue
			} else if len(srv) > 1 && kind == "terminating-gateway" {
				log.Errorf("too many service definitions found for: %s", service)
				continue
			}

			w.lock.Lock()
			if kind == "terminating-gateway" {
				w.id = srv[0].Service.ID
				w.name = srv[0].Service.Service
				w.address = srv[0].Service.Address
				w.port = srv[0].Service.Port
			} else {
				w.services[service].instances = srv
			}
			w.lock.Unlock()
			if dFirst && kind != "terminating-gateway" {
				w.services[service].ready.Wait()
				dFirst = false
			}
			w.notifyChanged()
		}
		if first {
			log.Infof("service config for %s ready", service)
			w.ready.Done()
			first = false
		}
	}
}

func (w *Watcher) handleProxyChange(first bool, gwServices *[]*api.GatewayService) {
	keep := make(map[string]bool)

	if gwServices != nil {
		for _, down := range *gwServices {
			keep[down.Service.Name] = true
			w.lock.Lock()
			_, ok := w.services[down.Service.Name]
			w.lock.Unlock()
			if !ok {
				if first {
					w.ready.Add(3)
				}
				w.startService(down, first)
			}
		}
	}

	for name := range w.services {
		if !keep[name] {
			w.removeService(name)
		}
	}
}

func (w *Watcher) startService(down *api.GatewayService, first bool) {

	d := &service{
		name:           down.Service.Name,
		gatewayService: down,
	}

	w.lock.Lock()
	w.services[down.Service.Name] = d
	w.lock.Unlock()

	d.ready.Add(2)
	go w.watchService(d.name, first, "")
	go w.watchLeaf(d.name, first)
	go w.watchIntention(d.name, first)
}

func (w *Watcher) removeService(name string) {
	log.Infof("removing downstream for service %s", name)

	w.lock.Lock()
	w.services[name].done = true
	delete(w.services, name)
	w.lock.Unlock()
	w.notifyChanged()
}

func (w *Watcher) watchCA() {
	log.Debugf("watching ca certs")

	first := true
	var lastIndex uint64
	for {
		caList, meta, err := w.consul.Agent().ConnectCARoots(&api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  10 * time.Minute,
		})
		if err != nil {
			log.Errorf("error fetching cas: %s", err)
			time.Sleep(errorWaitTime)
			if meta != nil {
				if meta.LastIndex < lastIndex || meta.LastIndex < 1 {
					lastIndex = 0
				}
			}
			continue
		}

		changed := lastIndex != meta.LastIndex
		lastIndex = meta.LastIndex

		if changed {
			log.Infof("CA certs changed, active root id: %s", caList.ActiveRootID)
			w.lock.Lock()
			w.certCAs = w.certCAs[:0]
			w.certCAPool = x509.NewCertPool()
			for _, ca := range caList.Roots {
				w.certCAs = append(w.certCAs, []byte(ca.RootCertPEM))
				ok := w.certCAPool.AppendCertsFromPEM([]byte(ca.RootCertPEM))
				if !ok {
					log.Warn("unable to add CA certificate to pool")
				}
			}
			w.lock.Unlock()
			w.notifyChanged()
		}

		if first {
			log.Infof("CA certs ready")
			w.ready.Done()
			first = false
		}
	}
}

func (w *Watcher) genCfg() Config {
	w.lock.Lock()

	defer func() {
		w.lock.Unlock()
		log.Debugf("done generating configuration")
	}()

	if len(w.services) == 0 {
		return Config{}
	}

	watcherConfig := Config{
		GatewayName:    w.name,
		GatewayID:      w.id,
		GatewayAddress: w.address,
		GatewayPort:    w.port,
		CAsPool:        w.certCAPool,
		CAs:            w.certCAs,
	}

	for _, down := range w.services {
		downstream := NewService(down)
		downstream.TLS.CAs = w.certCAs
		watcherConfig.Services = append(watcherConfig.Services, downstream)
	}
	return watcherConfig
}

func (w *Watcher) notifyChanged() {
	select {
	case w.update <- struct{}{}:
	default:
	}
}
