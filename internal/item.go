package internal

import (
	"encoding/xml"
	"io/ioutil"
	"sync"
)

type ItemInterfaceI interface {
	// add property
	Add(v interface{})
	Dumps() (String, error)
	Empty() bool
	Contains(string) bool
}

type MapItem struct {
	*Map
}


type ListItem struct {
	*List
}

type StringItem struct {
	String
	sync.RWMutex
}

func (s *StringItem) Add(v interface{}) {
	s.Lock()
	defer s.Unlock()
	switch v.(type) {
	case String:
		s.String = v.(String)
	}
}

func (s *StringItem) Dumps() (str String, err error) {
	str = s.String
	return
}

type Feeds struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *Channel `xml:"channel"`
}

func (f *Feeds) Add(v interface{}){
	switch v.(type) {
	case *XmlItem:
		f.Channel.AddItem(v.(*XmlItem))
	}
}

func (f *Feeds) Empty() bool{
	return len(f.Channel.Item) == 0
}

type Channel struct {
	Title         string  `xml:"title"`
	Link          string  `xml:"link"`
	Description   string  `xml:"description"`
	LastBuildDate string  `xml:"lastBuildDate"`
	Item          []*XmlItem `xml:"item"`
	sync.RWMutex
}

func (c *Channel) AddItem(item *XmlItem) {
	c.Lock()
	defer c.Unlock()
	c.Item = append(c.Item, item)
}

func (c *Channel) AddLastPubTime(pub string) {
	c.LastBuildDate = pub
}

func (f *Feeds) Dumps() (String, error){
	feeds, err := xml.Marshal(f)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("feeds.xml", feeds, 0666); err != nil {
		panic(err)
	}
	return String(feeds), err
}

type XmlItem struct {
	Title       string `xml:"title" validate:"required"`
	Link        string `xml:"link" validate:"required"`
	PubData     string `xml:"pubDate"`
	Description string `xml:"description" validate:"required"`
}

func NewFeeds() *Feeds {
	return &Feeds{
		XMLName: xml.Name{Local: "rss"},
		Version: "1.0.0",
		Channel: &Channel{
			Title:       "go-scrapy generater rss",
			Link:        "http://testerlife.com",
			Description: "RSS page automatically generated by spider",
			Item:        []*XmlItem{},
		},
	}
}
