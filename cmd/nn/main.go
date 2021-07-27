package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jhchabran/notionhacks"
	"github.com/urfave/cli/v2"
)

func getConfig(c *cli.Context) notionhacks.Config {
	if path := c.String("config"); path != "" {
		return notionhacks.NewJSONConfig(path)
	}
	return notionhacks.NewKeyChainConfig()
}

func main() {
	app := &cli.App{
		Name:     "nn",
		HelpName: "nn",
		Usage:    "Interact with notion.so databases from the command-line.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to json config file (debugging purposes)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "auth",
				Usage: "auth and store your api key in the keychain",
				Action: func(c *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Enter API key: ")
					text, _ := reader.ReadString('\n')
					config := getConfig(c)
					err := config.SetAPIKey(strings.TrimSpace(text))
					return err
				},
			},
			{
				Name:  "register-db",
				Usage: "register a database id by its name",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Required: true,
						Usage:    "name that will be used to refer to that database",
					},
					&cli.StringFlag{
						Name:     "id",
						Required: true,
						Usage:    "database-id",
					},
				},
				Action: func(c *cli.Context) error {
					config := getConfig(c)
					err := config.Load()
					if err != nil {
						return err
					}
					return config.RegisterDatabaseName(c.String("name"), c.String("id"))
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list items from a database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db",
						Required: true,
						Usage:    "name of the database to operate on",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "fancy",
						Usage: "output format, json, fancy or raw",
					},
				},
				Action: func(c *cli.Context) error {
					config := getConfig(c)
					err := config.Load()
					if err != nil {
						return err
					}
					client := notionhacks.New(config)
					items, raw, err := client.ListItems(c.String("db"))
					if err != nil {
						return err
					}
					if c.String("format") == "raw" {
						fmt.Println(string(raw))
					} else {
						for _, item := range items {
							fmt.Println(item.Name, item.Fields)
						}
					}
					return nil
				},
			},
			{
				Name:    "insert",
				Aliases: []string{"i", "a"},
				Usage:   "add an item to a database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db",
						Required: true,
						Usage:    "name of the database to operate on",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "item name (page title) to insert",
					},
					&cli.StringSliceFlag{
						Name:    "fields",
						Aliases: []string{"f"},
						Usage:   "list of fields to insert",
					},
				},
				Action: func(c *cli.Context) error {
					config := getConfig(c)
					err := config.Load()
					if err != nil {
						return err
					}
					fields, err := parseFields(c.StringSlice("fields"))
					if err != nil {
						return err
					}
					item := notionhacks.Item{
						Name:   c.String("name"),
						Fields: fields,
					}
					client := notionhacks.New(config)
					return client.InsertItem(c.String("db"), &item)
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e", "u"},
				Usage:   "edit an item from a database",
				Action: func(c *cli.Context) error {
					fmt.Println("list")
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFields(fields []string) (map[string]string, error) {
	m := map[string]string{}
	for _, field := range fields {
		strs := strings.Split(field, "=")
		if len(strs) < 2 {
			return nil, fmt.Errorf("invalid field format: %s", field)
		}
		key := strs[0]
		value := strings.Join(strs[1:], "")
		m[key] = value
	}
	return m, nil
}

// var headers = http.Header{
// 	"Authorization":  []string{"Bearer " + apiKey},
// 	"Notion-Version": []string{"2021-05-13"},
// 	"Content-Type":   []string{"application/json"},
// }

// func main() {
// 	client := &http.Client{
// 		Timeout: time.Second * 5,
// 	}

// 	u, _ := url.ParseRequestURI("https://api.notion.com/v1/pages")

// 	body := newTask("foobar", "PERSONAL", "Blog")

// 	req := http.Request{
// 		Method: "POST",
// 		URL:    u,
// 		Header: headers,
// 		Body:   io.NopCloser(body),
// 	}

// 	resp, err := client.Do(&req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		fmt.Println(resp.StatusCode)
// 		b, _ := ioutil.ReadAll(resp.Body)
// 		fmt.Println(string(b))
// 	}
// }

// func findProject(name string) (string, error) {
// 	client := &http.Client{
// 		Timeout: time.Second * 5,
// 	}

// 	u, _ := url.ParseRequestURI("https://api.notion.com/v1/databases/" + projectsDB + "/query")

// 	req := http.Request{
// 		Method: "POST",
// 		URL:    u,
// 		Header: headers,
// 	}

// 	resp, err := client.Do(&req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		fmt.Println(resp.StatusCode)
// 		b, _ := ioutil.ReadAll(resp.Body)
// 		fmt.Println(string(b))
// 	}

// 	r := map[string]interface{}{}
// 	dec := json.NewDecoder(resp.Body)
// 	err = dec.Decode(&r)
// 	if err != nil {
// 		return "", err
// 	}

// 	res := r["results"].([]interface{})
// 	for _, page := range res {
// 		m := page.(map[string]interface{})
// 		id := m["id"].(string)
// 		props := m["properties"].(map[string]interface{})
// 		nameProp := props["Name"].(map[string]interface{})
// 		fields := nameProp["title"].([]interface{})
// 		title := fields[0].(map[string]interface{})["plain_text"]

// 		if title == name {
// 			return id, nil
// 		}
// 	}

// 	return "", err
// }

// func newTask(title string, area string, project string) *bytes.Buffer {
// 	projectID, err := findProject("Blog")
// 	if err != nil {
// 		panic(err)
// 	}

// 	var buf bytes.Buffer

// 	m := map[string]interface{}{
// 		"parent": map[string]interface{}{
// 			"database_id": tasksDB,
// 		},
// 		"properties": map[string]interface{}{
// 			"Name": []interface{}{
// 				map[string]interface{}{
// 					"name": "Name",
// 					"text": map[string]interface{}{"content": title},
// 				},string
// 			},
// 			"Area": map[string]interface{}{
// 				"name": area,
// 			},
// 			"Status": map[string]interface{}{
// 				"name": "Inbox",
// 			},
// 			"Project": []interface{}{map[string]interface{}{
// 				"id": projectID,
// 			}},
// 		},
// 	}

// 	enc := json.NewEncoder(&buf)
// 	err = enc.Encode(m)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return &buf
// }
