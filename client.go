package notionhacks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jomei/notionapi"
)

type Client struct {
	notion *notionapi.Client
	config Config
}

func New(config Config) *Client {
	client := notionapi.NewClient(notionapi.Token(config.APIKey()))
	return &Client{
		notion: client,
		config: config,
	}
}

func pageTitle(page *notionapi.Page) (string, error) {
	if page == nil {
		return "", fmt.Errorf("cannot read title, nil page")
	}
	nameProp := page.Properties["Name"]
	if nameProp == nil {
		return "", fmt.Errorf("cannot read title, invalid property")
	}
	p, ok := nameProp.(*notionapi.TitleProperty)
	if !ok {
		return "", fmt.Errorf("cannot read title, not a page title property")
	}
	if len(p.Title) < 1 {
		return "", fmt.Errorf("cannot read title, invalid property")
	}
	return p.Title[0].PlainText, nil
}

func (c *Client) ListItems(dbname string) ([]*Item, []byte, error) {
	id, err := c.config.DatabaseID(dbname)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot list items, configuration error: %w", err)
	}
	resp, err := c.notion.Database.Query(context.Background(), notionapi.DatabaseID(id), &notionapi.DatabaseQueryRequest{
		Sorts: []notionapi.SortObject{},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot list items, request error: %w", err)
	}

	items := []*Item{}

	for _, p := range resp.Results {
		title, err := pageTitle(&p)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot list items: %w", err)
		}
		items = append(items, &Item{
			Name: title,
		})
	}

	return items, nil, nil
}

func (c *Client) InsertItem(dbname string, item *Item) (*notionapi.Page, error) {
	id, err := c.config.DatabaseID(dbname)
	if err != nil {
		return nil, fmt.Errorf("cannot insert item, configuration error: %w", err)
	}

	properties, err := c.collectProperties(notionapi.DatabaseID(id), item)
	if err != nil {
		return nil, fmt.Errorf("cannot insert item, error while collecting property configurations: %w", err)
	}

	properties["Name"] = notionapi.TitleProperty{
		Title: []notionapi.RichText{{Text: notionapi.Text{Content: item.Name}}},
	}

	page, err := c.notion.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: notionapi.DatabaseID(id),
		},
		Properties: properties,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot insert item: request error: %w", err)
	}
	return page, nil
}

func (c *Client) getPagePropertyTypes(ctx context.Context, id notionapi.DatabaseID) (notionapi.PropertyConfigs, error) {
	db, err := c.notion.Database.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return db.Properties, nil
}

func (c *Client) findPageByTitle(ctx context.Context, id notionapi.DatabaseID, title string) (*notionapi.Page, error) {
	resp, err := c.notion.Database.Query(ctx, id, &notionapi.DatabaseQueryRequest{
		PropertyFilter: &notionapi.PropertyFilter{
			Property: "Name",
			Text: &notionapi.TextFilterCondition{
				StartsWith: title,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	switch n := len(resp.Results); {
	case n == 0:
		return nil, fmt.Errorf("no page found with title starting with '%s'", title)
	case n == 1:
		return &resp.Results[0], nil
	default:
		return nil, fmt.Errorf("ambiguous title, %d pages found with title starting with '%s'", n, title)
	}
}

func (c *Client) collectProperties(id notionapi.DatabaseID, item *Item) (notionapi.Properties, error) {
	properties, err := c.getPagePropertyTypes(context.TODO(), notionapi.DatabaseID(id))
	if err != nil {
		return nil, err
	}

	// for k, v := range properties {
	// 	fmt.Println("name:", k, "value:", v)
	// }

	result := notionapi.Properties{}
	for k, v := range item.Fields {
		if p, ok := properties[k]; ok {
			switch p.GetType() {
			case "rich_text":
				result[k] = notionapi.RichTextProperty{RichText: []notionapi.RichText{{Text: notionapi.Text{Content: v}}}}
			case "number":
				f, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, err
				}
				result[k] = notionapi.NumberProperty{
					Type:   "number",
					Number: float64(f),
				}
			case "select":
				pp := p.(*notionapi.SelectPropertyConfig)
				var name string
				var found bool
				for _, option := range pp.Select.Options {
					// TODO case
					if option.Name == v {
						found = true
						name = option.Name
						break
					}
				}
				if found {
					result[k] = notionapi.SelectProperty{
						Type:   "select",
						Select: notionapi.Option{Name: name},
					}
				} else {
					return nil, fmt.Errorf("option not found")
				}
			case "multi_select":
				pp := p.(*notionapi.MultiSelectPropertyConfig)
				var name string
				var found bool
				for _, option := range pp.MultiSelect.Options {
					// TODO case
					if option.Name == v {
						found = true
						name = option.Name
						break
					}
				}
				if found {
					result[k] = notionapi.MultiSelectProperty{
						Type:        "multi_select",
						MultiSelect: []notionapi.Option{{Name: name}},
					}
				} else {
					return nil, fmt.Errorf("option not found")
				}
			case "date":
				// TODO ranges
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return nil, err
				}
				d := notionapi.Date(t)
				result[k] = notionapi.DateProperty{
					Type: "date",
					Date: notionapi.DateObject{
						Start: &d,
					},
				}
			case "formula":
				return nil, fmt.Errorf("formula property type is not supported")
			case "relation":
				pp := p.(*notionapi.RelationPropertyConfig)
				page, err := c.findPageByTitle(context.TODO(), pp.Relation.DatabaseID, v)
				if err != nil {
					return nil, err
				}
				result[k] = notionapi.RelationProperty{
					Type:     "relation",
					Relation: []notionapi.Relation{{ID: notionapi.PageID(page.ID)}},
				}
			case "rollup":
				return nil, fmt.Errorf("rollup property type is not supported")
			case "title":
				result[k] = notionapi.TitleProperty{
					Title: []notionapi.RichText{{Text: notionapi.Text{Content: v}}},
				}
			case "people":
				return nil, fmt.Errorf("TODO")
			case "files":
				return nil, fmt.Errorf("files property type is not supported")
			case "checkbox":
				b, err := strconv.ParseBool(v)
				if err != nil {
					return nil, err
				}
				result[k] = notionapi.CheckboxProperty{
					Type:     "checkbox",
					Checkbox: b,
				}
			case "url":
				result[k] = notionapi.URLProperty{
					Type: "url",
					URL:  v,
				}
			case "email":
				result[k] = notionapi.EmailProperty{
					Type:  "email",
					Email: v,
				}
			case "phone_number":
				result[k] = notionapi.PhoneNumberProperty{
					Type:        "phone_number",
					PhoneNumber: v,
				}
			default:
				return nil, fmt.Errorf("read only property %s", k)
			}
		}
	}
	return result, nil
}
