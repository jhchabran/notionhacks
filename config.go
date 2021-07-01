package notionhacks

import (
	"fmt"
	"strings"

	"github.com/99designs/keyring"
)

func SaveApiKey(apiKey string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "notionhacks",
	})
	if err != nil {
		return err
	}
	err = ring.Set(keyring.Item{
		Key:  "API_KEY",
		Data: []byte(apiKey),
	})
	if err != nil {
		return err
	}
	return nil
}

type Config struct {
	ring      keyring.Keyring
	apiKey    string
	databases map[string]databaseID
}

func Load() (*Config, error) {
	c := Config{}
	err := c.Load()
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Load() error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "notionhacks",
	})
	if err != nil {
		return err
	}
	c.ring = ring

	v, err := ring.Get("API_KEY")
	if err != nil {
		return err
	}
	c.apiKey = string(v.Data)

	keys, err := ring.Keys()
	if err != nil {
		return err
	}

	c.databases = map[string]databaseID{}
	for _, k := range keys {
		if strings.HasPrefix(k, "DB_") {
			name := strings.TrimPrefix(k, "DB_")
			v, err := ring.Get(k)
			if err != nil {
				return err
			}
			c.databases[strings.ToLower(name)] = databaseID(v.Data)
		}
	}

	return nil
}

func (c *Config) RegisterDatabase(name string, ID string) {
	c.databases[name] = databaseID(ID)
}

func (c *Config) Save() error {
	for k, v := range c.databases {
		err := c.ring.Set(keyring.Item{
			Key:  fmt.Sprintf("DB_%s", strings.ToUpper(k)),
			Data: []byte(v),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
