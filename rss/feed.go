package rss

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"
)

type Torrent struct {
	Title       string    `json:"title"`
	GUID        string    `json:"guid"`
	PubDate     time.Time `json:"pubDate,omitempty"`
	Size        string    `json:"size,omitempty"`
	MagnetURI   string    `json:"magnetURI,omitempty"`
	TorrentFile string    `json:"torrentFile,omitempty"`
}

type Feed struct {
	URL       *url.URL
	exitCh    chan os.Signal
	recordCh  chan *Torrent
	refreshCh chan struct{}
	rssParser *gofeed.Parser
}

func NewRSSFeed(feedurl string, exitCh chan os.Signal, recordCh chan *Torrent, refreshCh chan struct{}) *Feed {
	u, err := url.ParseRequestURI(feedurl)
	if err != nil {
		log.Printf("[ERROR] Unable to parse feed url: %s", feedurl)
		exitCh <- syscall.SIGHUP
	}

	feed := &Feed{
		URL:       u,
		exitCh:    exitCh,
		recordCh:  recordCh,
		refreshCh: refreshCh,
		rssParser: gofeed.NewParser(),
	}
	return feed
}

func (f *Feed) Fetch() {
	for {
		<-f.refreshCh
		log.Println("[INFO] Refresh action triggered")
		rss, err := f.rssParser.ParseURL(f.URL.String())
		if err != nil {
			log.Printf("[ERROR] Unable to parse feed: %s", err)
			f.exitCh <- syscall.SIGHUP
		}

		for _, item := range rss.Items {
			size, err := strconv.Atoi(item.Enclosures[0].Length)
			if err != nil {
				log.Printf("[ERROR] Unable to parse the media size")
				continue
			}
			size = size / 1048576
			f.recordCh <- &Torrent{
				Title:       item.Title,
				GUID:        item.GUID,
				PubDate:     *item.PublishedParsed,
				Size:        fmt.Sprintf("%dmb", size),
				TorrentFile: item.Enclosures[0].URL,
				MagnetURI:   item.Extensions["torrent"]["magnetURI"][0].Value,
			}
		}
	}
}
