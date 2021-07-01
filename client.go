package notionhacks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const endpoint = "https://api.notion.com/v1"

type databaseID string

type Client struct {
	config *Config
	header http.Header
}

func New(config *Config) *Client {
	header := http.Header{
		"Authorization":  []string{"Bearer " + config.apiKey},
		"Notion-Version": []string{"2021-05-13"},
		"Content-Type":   []string{"application/json"},
	}

	return &Client{
		config: config,
		header: header,
	}
}

func (c *Client) newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func (c *Client) newRequest(method string, path string) *http.Request {
	u, _ := url.ParseRequestURI(endpoint + path)
	return &http.Request{
		Method: method,
		URL:    u,
		Header: c.header,
	}
}

func (c *Client) InsertItem(db string, item *Item) error {
	id, ok := c.config.databases[db]
	if !ok {
		return fmt.Errorf("unknown database name: %s", db)
	}
	return c.insertItem(id, item)
}

func (c *Client) insertItem(db databaseID, item *Item) error {
	var buf bytes.Buffer

	m := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": db,
		},
		"properties": map[string]interface{}{
			"Name": []interface{}{
				map[string]interface{}{
					"name": "Name",
					"text": map[string]interface{}{"content": item.Name},
				},
			},
		},
	}

	props := m["properties"].(map[string]interface{})
	for k, v := range item.Fields {
		props[k] = map[string]interface{}{"name": v}
	}

	enc := json.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return err
	}

	fmt.Println(m)

	cl := c.newHTTPClient()
	req := c.newRequest("POST", "/pages")
	req.Body = io.NopCloser(&buf)
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		b, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(b))
		return fmt.Errorf("failed to perform request")
	}
	return nil
}
