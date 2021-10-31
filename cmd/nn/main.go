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
					err := config.Load()
					if err != nil {
						return err
					}
					err = config.SetAPIKey(strings.TrimSpace(text))
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
					page, err := client.InsertItem(c.String("db"), &item)
					if err != nil {
						return err
					}
					fmt.Println(page.URL)
					return nil
				},
			},
			{
				Name:    "dump",
				Aliases: []string{"d"},
				Usage:   "Dump a page in markdown format",
				Action: func(c *cli.Context) error {
					config := getConfig(c)
					err := config.Load()
					if err != nil {
						return err
					}
					client := notionhacks.New(config)
					return notionhacks.Dump(client, c.Args().First())
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
