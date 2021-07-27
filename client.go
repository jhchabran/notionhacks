package notionhacks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jomei/notionapi"
)

const endpoint = "https://api.notion.com/v1"

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
	p, ok := nameProp.(*notionapi.PageTitleProperty)
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
		return nil, nil, err
	}
	resp, err := c.notion.Database.Query(context.Background(), notionapi.DatabaseID(id), &notionapi.DatabaseQueryRequest{
		Sorts: []notionapi.SortObject{},
	})
	if err != nil {
		return nil, nil, err
	}

	items := []*Item{}

	for _, p := range resp.Results {
		title, err := pageTitle(&p)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &Item{
			Name: title,
		})
	}

	return items, nil, nil
}

func (c *Client) InsertItem(dbname string, item *Item) error {
	id, err := c.config.DatabaseID(dbname)
	if err != nil {
		return err
	}

	dbprops, err := c.getPagePropertyTypes(context.TODO(), notionapi.DatabaseID(id))
	properties, err := collectProperties(item, dbprops)
	if err != nil {
		return err
	}

	properties["Name"] = notionapi.PageTitleProperty{
		Title: notionapi.Paragraph{notionapi.RichText{Text: notionapi.Text{Content: item.Name}}},
	}

	_, err = c.notion.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: notionapi.DatabaseID(id),
		},
		Properties: properties,
	})
	return err
}

func (c *Client) getPagePropertyTypes(ctx context.Context, id notionapi.DatabaseID) (notionapi.Properties, error) {
	db, err := c.notion.Database.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return db.Properties, nil
}

func collectProperties(item *Item, properties notionapi.Properties) (notionapi.Properties, error) {
	result := notionapi.Properties{}
	for k, v := range item.Fields {
		if p, ok := properties[k]; ok {
			switch p.GetType() {
			case "rich_text":
				pp := p.(*notionapi.RichTextProperty)
				pp.RichText = []notionapi.RichText{{Text: notionapi.Text{Content: v}}}
				result[k] = pp
			case "number":
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				result[k] = notionapi.NumberValueProperty{
					Type:   "number",
					Number: f,
				}
			case "select":
				pp := p.(*notionapi.SelectProperty)
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
					result[k] = notionapi.SelectOptionProperty{
						Type:   "select",
						Select: notionapi.Option{Name: name},
					}
				} else {
					return nil, fmt.Errorf("option not found")
				}
			case "multi_select":
				pp := p.(*notionapi.MultiSelectProperty)
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
					result[k] = notionapi.MultiSelectOptionsProperty{
						Type:        "multi_select",
						MultiSelect: []notionapi.Option{notionapi.Option{Name: name}},
					}
				} else {
					return nil, fmt.Errorf("option not found")
				}
			case "date":
				// TODO ranges
				_, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return nil, err
				}
				result[k] = notionapi.DateProperty{
					Type: "date",
					Date: map[string]string{"start": v},
				}
			case "formula":
				return nil, fmt.Errorf("formula property type is not supported")
			case "relation":
				return nil, fmt.Errorf("TODO")
			case "rollup":
				return nil, fmt.Errorf("rollup property type is not supported")
			case "title":
				result[k] = notionapi.PageTitleProperty{
					Title: notionapi.Paragraph{notionapi.RichText{Text: notionapi.Text{Content: v}}},
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
