package config

import (
	"fmt"
	"strings"

	"github.com/f5devcentral/bigip-tgw/as3"
	"github.com/f5devcentral/bigip-tgw/consul"
	"github.com/spf13/viper"
)

//DEFAULTS
var (
	defaultSchema        string   = "https://raw.githubusercontent.com/F5Networks/f5-appsvcs-extension/master/schema/latest/as3-schema.json"
	defaultSchemaVersion string   = "3.20.0"
	defaultUsername      string   = "admin"
	defaultPort          string   = "8443"
	requiredKeys         []string = []string{"gateway.name", "bigip.bigipurl", "bigip.bigippassword"}
)

type Config struct {
	Gateway GatewayConfig
	Bigip   as3.Params
	Consul  consul.ConsulConfig
}

type GatewayConfig struct {
	Name      string
	Namespace string
}

/*
// Struct of potential configurable options
type Params struct {
	// Package local for unit testing only
	SchemaLocal               string
	AS3Validation             bool
	SSLInsecure               bool
	EnableTLS                 string
	TLS13CipherGroupReference string
	Ciphers                   string
	//Agent                     string
	OverriderCfgMapName string
	SchemaLocalPath     string
	FilterTenants       bool
	BIGIPUsername       string
	BIGIPPassword       string
	BIGIPURL            string
	TrustedCerts        string
	AS3PostDelay        int
	//ConfigWriter        writer.Writer
	EventChan chan interface{}
	//Log the AS3 response body in Controller logs
	LogResponse               bool
	RspChan                   chan interface{}
	UserAgent                 string
	As3Version                string
	As3Release                string
	unprocessableEntityStatus bool
}
*/

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	v.SetDefault("bigip.schema", defaultSchema)
	v.SetDefault("bigip.schemaversion", defaultSchemaVersion)
	v.SetDefault("bigip.BIGIPUsername", defaultUsername)
	//v.SetDefault("bigip.port", defaultPort)

	v.BindEnv("consul.address")
	v.BindEnv("consul.token")
	v.BindEnv("consul.namespace")

	v.BindEnv("bigip.BIGIPURL")
	v.BindEnv("bigip.BIGIPUsername")
	v.BindEnv("bigip.BIGIPPassword")

	v.BindEnv("gateway.name")
	c := &Config{
		Gateway: GatewayConfig{},
		Bigip:   as3.Params{},
		Consul:  consul.ConsulConfig{},
	}
	err = v.Unmarshal(c)
	if err != nil {
		return c, err
	}
	for _, key := range requiredKeys {
		if v.Get(key) == nil {
			return c, fmt.Errorf("configuration element %s is not set", key)
		}
	}
	return c, err
}
