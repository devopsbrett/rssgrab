package es

import (
	"log"
	"os"
	"syscall"

	"github.com/devopsbrett/rssgrab/rss"
	"gopkg.in/olivere/elastic.v3"
)

type Client struct {
	Conn      *elastic.Client
	indexName string
	typeName  string
	exitCh    chan os.Signal
	recordCh  chan *rss.Torrent
}

func NewClient(url, indexName, typeName string, exitCh chan os.Signal, recordCh chan *rss.Torrent) *Client {
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		log.Printf("[ERROR] Unable to connect to elasticsearch: %s", err)
		exitCh <- syscall.SIGHUP
		return &Client{}
	}
	exists, err := client.IndexExists(indexName).Do()
	if err != nil {
		log.Printf("[ERROR] Error while checking if index exists: %s", err)
		exitCh <- syscall.SIGHUP
		return &Client{}
	}

	if !exists {
		log.Println("[INFO] Creating index")
		_, err := client.CreateIndex(indexName).Do()
		if err != nil {
			log.Printf("[ERROR] Error while creating index: %s", err)
			exitCh <- syscall.SIGHUP
			return &Client{}
		}
	}

	return &Client{
		Conn:      client,
		indexName: indexName,
		typeName:  typeName,
		recordCh:  recordCh,
		exitCh:    exitCh,
	}
}

func (c *Client) Store() {
	for {
		torrent := <-c.recordCh
		c.Conn.Index().Index(c.indexName).Type(c.typeName).Id(torrent.GUID).BodyJson(torrent).Do()
		c.Conn.Flush().Index(c.indexName).Do()

		log.Println("[INFO] Storing", torrent.Title)
	}
}
