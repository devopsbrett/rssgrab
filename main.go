package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devopsbrett/rssgrab/es"
	"github.com/devopsbrett/rssgrab/rss"
)

type Torrent struct {
	Title       string    `json:"title"`
	PubDate     time.Time `json:"pubDate,omitempty"`
	Size        string    `json:"size,omitempty"`
	MagnetURI   string    `json:"magnetURI,omitempty"`
	TorrentFile string    `json:"torrentFile,omitempty"`
}

type Settings struct {
	rssURL  string
	esURL   string
	esIndex string
	esType  string
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	set := new(Settings)
	flag.StringVar(&set.rssURL, "rss", "https://eztv.immunicity.eu/ezrss.xml", "URL of rss feed")
	flag.StringVar(&set.esURL, "es", "http://127.0.0.1:9200", "URL of elasticsearch")
	flag.StringVar(&set.esIndex, "index", "torrents", "Name of ElasticSearch index to use")
	flag.StringVar(&set.esType, "type", "show", "Name of ElasticSearch type to use")

	flag.Parse()

	exitCh := make(chan os.Signal, 2)
	recordCh := make(chan *rss.Torrent, 30)
	refreshCh := make(chan struct{})
	signal.Notify(exitCh, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	rssFeed := rss.NewRSSFeed(set.rssURL, exitCh, recordCh, refreshCh)
	esClient := es.NewClient(set.esURL, set.esIndex, set.esType, exitCh, recordCh)

	go rssFeed.Fetch()
	go esClient.Store()

	refreshCh <- struct{}{}

	for {
		select {
		case s := <-exitCh:
			log.Printf("Receive on exitCh")
			switch s {
			case os.Interrupt:
				return 2
			case syscall.SIGHUP:
				return 9
			case syscall.SIGTERM:
				return 255
			}
		case <-time.After(15 * time.Minute):
			refreshCh <- struct{}{}
		}
	}

	// fp := gofeed.NewParser()
	// feed, _ := fp.ParseURL(set.rssURL)
	// client, err := elastic.NewClient(elastic.SetURL(set.esURL))
	// if err != nil {
	// 	panic(err)
	// }

	// exists, err := client.IndexExists(set.esIndex).Do()
	// if err != nil {
	// 	panic(err)
	// }

	// if !exists {
	// 	fmt.Println("Creating index")
	// 	createIndex, err := client.CreateIndex(set.esIndex).Do()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if !createIndex.Acknowledged {
	// 		fmt.Println("Create index wasn't acknowledged")
	// 	}
	// } else {
	// 	fmt.Println("Index already exists")
	// }
	// var tor Torrent
	// var size int
	// for _, item := range feed.Items {
	// 	s, _ := strconv.Atoi(item.Enclosures[0].Length)
	// 	size = s / 1048576
	// 	tor = Torrent{
	// 		Title:       item.Title,
	// 		PubDate:     *item.PublishedParsed,
	// 		Size:        fmt.Sprintf("%dmb", size),
	// 		TorrentFile: item.Enclosures[0].URL,
	// 		MagnetURI:   item.Extensions["torrent"]["magnetURI"][0].Value,
	// 	}
	// 	_, err := client.Index().Index(set.esIndex).Type(set.esType).Id(item.GUID).BodyJson(tor).Do()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// if err != nil {
	// 	fmt.Println("Couldn't connect")
	// 	panic(err)
	// }
	// info, code, err := client.Ping(set.esURL).Do()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Elasticsearch returned with code %d and version %s", code, info.Version.Number)
}
