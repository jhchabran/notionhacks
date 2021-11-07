package notionhacks

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
)

type Config interface {
	Load() error
	SetAPIKey(apiKey string) error
	APIKey() string
	DatabaseID(dbname string) (string, error)
	RegisterDatabaseName(name string, ID string) error
	ListDatabases() []string
}

type JSONConfig struct {
	path      string
	ApiKey    string            `json:"api_key"`
	Databases map[string]string `json:"databases"`
}

func NewJSONConfig(path string) *JSONConfig {
	return &JSONConfig{path: path}
}

func (c *JSONConfig) Load() error {
	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(f)
	return dec.Decode(c)
}

func (c *JSONConfig) RegisterDatabaseName(name string, id string) error {
	return fmt.Errorf("not implemented")
}

func (c *JSONConfig) SetAPIKey(key string) error {
	return fmt.Errorf("not implemented")
}

func (c *JSONConfig) APIKey() string {
	return c.ApiKey
}

func (c *JSONConfig) DatabaseID(dbname string) (string, error) {
	id, ok := c.Databases[dbname]
	if !ok {
		return "", fmt.Errorf("no database registered with the name: %s", dbname)
	}
	return id, nil
}

func (c *JSONConfig) ListDatabases() []string {
	dbs := make([]string, 0, len(c.Databases))
	for name := range c.Databases {
		dbs = append(dbs, name)
	}
	return dbs
}

type KeyChainConfig struct {
	ring      keyring.Keyring
	apiKey    string
	databases map[string]string
}

func NewKeyChainConfig() *KeyChainConfig {
	c := KeyChainConfig{}
	return &c
}

func (c *KeyChainConfig) SetAPIKey(apiKey string) error {
	c.apiKey = apiKey
	err := c.ring.Set(keyring.Item{
		Key:  "API_KEY",
		Data: []byte(apiKey),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *KeyChainConfig) APIKey() string {
	return c.apiKey
}

func (c *KeyChainConfig) Load() error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "notionhacks",
	})
	if err != nil {
		return err
	}
	c.ring = ring

	v, err := ring.Get("API_KEY")
	if err != nil && err != keyring.ErrKeyNotFound {
		return err
	}
	c.apiKey = string(v.Data)

	keys, err := ring.Keys()
	if err != nil {
		return err
	}

	c.databases = map[string]string{}
	for _, k := range keys {
		if strings.HasPrefix(k, "DB_") {
			name := strings.TrimPrefix(k, "DB_")
			v, err := ring.Get(k)
			if err != nil {
				return err
			}
			c.databases[strings.ToLower(name)] = string(v.Data)
		}
	}

	return nil
}

func (c *KeyChainConfig) DatabaseID(dbname string) (string, error) {
	id, ok := c.databases[dbname]
	if !ok {
		return "", fmt.Errorf("no database registered with the name: %s", dbname)
	}
	return id, nil
}

func (c *KeyChainConfig) RegisterDatabaseName(name string, ID string) error {
	c.databases[name] = string(ID)
	return c.saveDBs()
}

func (c *KeyChainConfig) ListDatabases() []string {
	dbs := make([]string, 0, len(c.databases))
	for name := range c.databases {
		dbs = append(dbs, name)
	}
	return dbs
}

func (c *KeyChainConfig) saveDBs() error {
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
