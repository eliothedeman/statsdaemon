package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"github.com/eliothedeman/statsdaemon/backend"
)

const (
	DEFAULT_SERVICE_ADDRESS    = ":8125"
	DEFAULT_FLUSH_INTERVAL     = time.Second * 10
	DEFAULT_DEBUG              = false
	DEFUALT_PERSIST_COUNT_KEYS = 60
	DEFUALT_PREFIX             = ""
)

var (
	WRONG_CONFIG_TYPE = errors.New("wrong config type")
)

type Config struct {
	ServiceAddress      string                     `json:"service_address"`
	FlushInterval       time.Duration              `json:"-"`
	FlushIntervalString string                     `json:"flush_interval"`
	Debug               bool                       `json:"debug"`
	PersistCountKeys    int64                      `json:"persist_count_keys"`
	Prefix              string                     `json:"prefix"`
	Backends            map[string]json.RawMessage `json:"backends"`
}

// apply the config to all of the loaded backends, and return all of the configured backends
func (c *Config) InitBackends() ([]backend.Backend, error) {
	loaded := make([]backend.Backend, 0, len(c.Backends))
	for k, raw := range c.Backends {
		v, inMap := backend.Backends[k]
		if inMap {
			conf := v.ConfigStruct()
			err := json.Unmarshal(raw, conf)
			if err != nil {
				return loaded, err
			}
			err = v.Init(conf)
			if err != nil {
				return loaded, err
			}
			log.Printf("Loaded config for %s\n", k)
			loaded = append(loaded, v)
		} else {
			log.Println("No backend found for ", k)
		}
	}
	return loaded, nil
}

// Given a file name, load the config that resides at that path
func LoadConfigFromFile(name string) (*Config, error) {
	c := &Config{
		ServiceAddress:   DEFAULT_SERVICE_ADDRESS,
		FlushInterval:    DEFAULT_FLUSH_INTERVAL,
		Debug:            DEFAULT_DEBUG,
		PersistCountKeys: DEFUALT_PERSIST_COUNT_KEYS,
		Prefix:           DEFUALT_PREFIX,
		Backends:         make(map[string]json.RawMessage),
	}

	buff, err := ioutil.ReadFile(name)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal(buff, c)
	if err != nil {
		log.Fatal(err)
	}

	if c.FlushIntervalString != "" {
		c.FlushInterval, err = time.ParseDuration(c.FlushIntervalString)
	}
	if c.Debug {
		log.SetFlags(log.Llongfile)
	}
	return c, err
}
