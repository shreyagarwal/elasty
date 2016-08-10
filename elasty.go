package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lunux2008/xulu"
	"github.com/streadway/amqp"
	"github.com/urfave/cli"
)

type esBulkCntrlStmnt struct {
	_index             string `json:"_index,omitempty"`
	_type              string `json:"_type,omitempty"`
	_id                string `json:"_id,omitempty"`
	_retry_on_conflict string `json:"_retry_on_conflict,omitempty"`
	_version           string `json:"_version,omitempty"`
}

type esBulkStmntType struct {
	s_delete esBulkCntrlStmnt `json:"delete"`
	// s_create esBulkCntrlStmnt `json:"create,omitempty"`
	// s_insert esBulkCntrlStmnt `json:"insert,omitempty"`
	// s_update esBulkCntrlStmnt `json:"update,omitempty"`
}

func main() {

	// processRaw(`{ "delete": { "_index": "website", "_type": "blog", "_id": "123" }}
	//        { "create": { "_index": "website", "_type": "blog", "_id": "123" }}
	//        { "title":    "My first blog post" }
	//        { "index":  { "_index": "website", "_type": "blog" }}
	//        { "title":    "My second blog post" }
	//        { "update": { "_index": "website", "_type": "blog", "_id": "123", "_retry_on_conflict" : 3}}
	//        { "doc" : {"title" : "My updated blog post"}}
	//        `)

	var url string
	xulu.Use(url)

	app := cli.NewApp()
	app.Name = "elasty"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Shrey Agarwal",
			Email: "s@shreyagarwal.com",
		},
	}
	app.Usage = "Elasticsearch toolbelt based on experience"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url",
			Value:       "http://localhost:9200",
			Usage:       "connect url stub",
			Destination: &url,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "threadpool",
			Aliases: []string{"th"},
			Usage:   "Show cluster threadpool",
			Action: func(c *cli.Context) error {

				return nil
			},
		},
		{
			Name:    "rmqtoes",
			Aliases: []string{"c"},
			Usage:   "complete a task on the list",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "options for task templates",
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add a new template",
					Action: func(c *cli.Context) error {
						fmt.Println("new task template: ", c.Args().First())
						return nil
					},
				},
				{
					Name:  "remove",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) error {
						fmt.Println("removed task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)

}

func processRaw(rawData string) {

	/*
	   - Split each string
	   - Json parse each string
	   - Classify the operation
	   - Count number of operations and number of lines
	   - Add lines in buffer ( Check thread concurrency here )
	   - Take a call whether to insert data in ES or not
	       - Lock the data and insert it and empty it
	*/

	// split the data
	// fmt.Printf("%d\n", len(strings.Split(rawData, "\n")))

	var splits []string = strings.Split(rawData, "\n")
	xulu.Use(splits)

	// work on each split
	fmt.Printf("Splitting in lines : %d\n", len(splits))

	fmt.Printf("Marshalling Line : %q\n", splits[0])

	// Json unmarshal
	// var stmnt1 esBulkStmntType
	// err := json.Unmarshal([]byte(splits[0]), &stmnt1)
	// if err != nil {
	// 	// panic(err)
	// 	log.Fatalf("json.Unmarshal: %s", err)
	// }

	// var parsed interface{}
	var parsed map[string]interface{}

	err := json.Unmarshal([]byte(splits[0]), &parsed)
	if err != nil {
		panic(err)
		log.Fatalf("json.Unmarshal interface: %s", err)
	}

	remarsh2, _ := json.Marshal(parsed)

	// map[delete:map[_id:123 _index:website _type:blog]]
	fmt.Println(string(remarsh2))
	// fmt.Println("%q\n", parsed, map[delete)

	// fmt.Printf("%q\n", parsed.(type))
}

func parseThreadPoolOutpu(bulkData string) {
	/*
	   t := "id   pid   ip        host      bulk.active bulk.queue"
	   fmt.Printf("%q\n", strings.Fields(t))
	*/

}

func hitEs(bulkData string) {
	resp, err := http.Get("http://localhost:9200/_cat/thread_pool?v&h=id,pid,ip,host,bulk.active,bulk.queue")

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ThreadPool, %q", body)

}

// Initizlise rabbit Mq connection
func initialize() {

	// Reconnect on conection Close
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Println("reconnect after 5 seconds: " + err.Error())

		time.Sleep(5 * 1000 * time.Millisecond)

		initialize()
	}()

	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connection.open: %s", err)
		panic("cannot connect")
	}
	conn.NotifyClose(c)
	// defer conn.Close()
	fmt.Printf("Connection open\n")

	// Opening channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel.open: %s", err)
		panic("cannot open channel")
	}
	ch.NotifyClose(c)
	fmt.Printf("Channel open\n")

	// Declare exchange
	err = ch.ExchangeDeclare("go_ex1", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("channel.ExchangeDeclare: %s", err)
		panic("cannot declare exchange")
	}
	fmt.Printf("Exchange configured\n")

	// declare Queue
	q, err := ch.QueueDeclare("go_q1", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("channel.QueueDeclare: %s", err)
		panic("cannot declare q")
	}
	fmt.Printf("Q configured\n")
	_ = q

	// Q bind
	err = ch.QueueBind("go_q1", "#", "go_ex1", false, nil)
	if err != nil {
		log.Fatalf("channel.QueueBind: %s", err)
		panic("cannot QueueBind")
	}
	fmt.Printf("Q bound\n")

	// Qos

	// Set our quality of service.  Since we're sharing 3 consumers on the same
	// channel, we want at least 3 messages in flight.
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("basic.qos: %v", err)
	}

	//consume
	test_msgs, err := ch.Consume("go_q1", "go_q1_CONSUMER", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("basic.consume: %v", err)
	}

	go func() {
		for each_msg := range test_msgs {
			log.Println("Msg: %s", string(each_msg.Body[:]))

			err = each_msg.Ack(false)
			if err != nil {
				log.Fatalf("Error in Ack: %v", err)
			}
		}
	}()

	// send it to Elasticsearch as soon as you receive it .. and wait on receiving

}
