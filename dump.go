package notionhacks

import (
	"context"
	"fmt"
	"strings"

	"github.com/jomei/notionapi"
)

func WatcherStart(ctx context.Context) error {
	return nil
}

func Dump(c *Client, id string) error {
	blocks, err := c.FindPageBlocks(context.TODO(), notionapi.BlockID(id))

	if err != nil {
		return err
	}

	var iw IndentWriter
	for _, r := range blocks {
		err := AppendMarkdown(&iw, r)
		if err != nil {
			return err
		}
	}

	fmt.Println(iw.String())
	return nil
}

type RichText notionapi.RichText

func (rt RichText) AppendMarkdown(iw *IndentWriter) error {
	if rt.Annotations == nil {
		return iw.Append(rt.PlainText)
	}

	surrounds := []string{}
	if rt.Annotations.Italic {
		surrounds = append(surrounds, "_")
	}
	if rt.Annotations.Bold {
		surrounds = append(surrounds, "*")
	}
	if rt.Annotations.Code {
		surrounds = []string{"`"}
	}
	begin := strings.Join(surrounds, "")
	var end string
	for _, c := range surrounds {
		end = c + end
	}
	res := begin + rt.PlainText + end

	if rt.Href != "" {
		res = "(" + res + ")[" + rt.Href + "]"
	}

	return iw.Append(res)
}

type RichTexts []notionapi.RichText

func (rts RichTexts) AppendMarkdown(iw *IndentWriter) error {
	for _, rt := range rts {
		err := RichText(rt).AppendMarkdown(iw)
		if err != nil {
			return err
		}
	}
	return nil
}

type Heading1 notionapi.Heading1Block

func (h1 Heading1) AppendMarkdown(iw *IndentWriter) error {
	err := iw.Append("# ")
	if err != nil {
		return err
	}
	err = RichTexts(h1.Heading1.Text).AppendMarkdown(iw)
	if err != nil {
		return err
	}
	iw.CR()
	return nil
}

type Heading2 notionapi.Heading2Block

func (h2 Heading2) AppendMarkdown(iw *IndentWriter) error {
	err := iw.Append("## ")
	if err != nil {
		return err
	}
	err = RichTexts(h2.Heading2.Text).AppendMarkdown(iw)
	if err != nil {
		return err
	}
	iw.CR()
	return nil
}

type Heading3 notionapi.Heading3Block

func (h3 Heading3) AppendMarkdown(iw *IndentWriter) error {
	err := iw.Append("### ")
	if err != nil {
		return err
	}
	err = RichTexts(h3.Heading3.Text).AppendMarkdown(iw)
	if err != nil {
		return err
	}
	iw.CR()
	return nil
}

type BullettedListItem notionapi.BulletedListItemBlock

type IndentWriter struct {
	sb     strings.Builder
	indent int
	dirty  bool
}

func (mw *IndentWriter) Append(s string) error {
	if !mw.dirty {
		mw.dirty = true
		for i := 0; i < mw.indent; i++ {
			_, err := mw.sb.WriteRune(' ')
			if err != nil {
				return err
			}
			_, err = mw.sb.WriteRune(' ')
			if err != nil {
				return err
			}
		}
	}
	_, err := mw.sb.WriteString(s)
	return err
}

func (mw *IndentWriter) Indent() {
	mw.indent++
}

func (mw *IndentWriter) Outdent() {
	mw.indent--
}

func (mw *IndentWriter) String() string {
	return mw.sb.String()
}

func (mw *IndentWriter) CR() {
	mw.dirty = false
	_, _ = mw.sb.WriteRune('\n')
}

func (bi BullettedListItem) AppendMarkdown(iw *IndentWriter) error {
	err := iw.Append("- ")
	if err != nil {
		return err
	}
	err = RichTexts(bi.BulletedListItem.Text).AppendMarkdown(iw)
	if err != nil {
		return err
	}
	iw.CR()
	if !bi.HasChildren {
		return nil
	}

	for _, c := range bi.BulletedListItem.Children {
		iw.Indent()
		err := AppendMarkdown(iw, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func AppendMarkdown(iw *IndentWriter, block notionapi.Block) error {
	switch bt := block.GetType(); bt {
	case notionapi.BlockTypeHeading1:
		b := block.(*notionapi.Heading1Block)
		err := Heading1(*b).AppendMarkdown(iw)
		return err
	case notionapi.BlockTypeHeading2:
		b := block.(*notionapi.Heading2Block)
		err := Heading2(*b).AppendMarkdown(iw)
		return err
	case notionapi.BlockTypeHeading3:
		b := block.(*notionapi.Heading3Block)
		err := Heading3(*b).AppendMarkdown(iw)
		return err
	case notionapi.BlockTypeParagraph:
		b := block.(*notionapi.ParagraphBlock)
		err := RichTexts(b.Paragraph.Text).AppendMarkdown(iw)
		// TODO
		iw.CR()
		return err
	case notionapi.BlockTypeBulletedListItem:
		b := block.(*notionapi.BulletedListItemBlock)
		err := BullettedListItem(*b).AppendMarkdown(iw)
		return err
	default:
		fmt.Println(bt)
	}
	return nil
}
