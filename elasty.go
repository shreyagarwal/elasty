package main

import (
	"bytes"
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

/* GLobal variables */

// rabbit mq variables
var esUrl, esIndex, rmqConnectStr, exName, exKind, qName, qBindKey string
var rmqReconnTimeout int
var dryrun bool = false

// Buffer variables global
// var bufMsgs byte[]
// var bufMsgCount int

/* End global variables */

func main() {

	/* Main function only has CLI parsing */
	cliArgsParse()
}

/* Parse the CLI args and call appropriate function*/
func cliArgsParse() {

	xulu.Use(esUrl, esIndex)

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
			Name:        "url, u",
			Value:       "http://localhost:9200",
			Usage:       "connect url stub ( default http://localhost:9200 )",
			Destination: &esUrl,
		},
		cli.StringFlag{
			Name:        "index, i",
			Value:       "test",
			Usage:       "index name ( default test )",
			Destination: &esIndex,
		},

		// Rmq2ES flags
		cli.StringFlag{
			Name:        "rmqconnectstr, r",
			Value:       "amqp://guest:guest@localhost:5672/",
			Usage:       "For rmq2es : RabbitMq Connection String ( default amqp://guest:guest@localhost:5672/ )",
			Destination: &rmqConnectStr,
		},
		cli.IntFlag{
			Name:        "rmqreconntimeout",
			Value:       5 * 1000,
			Usage:       "For rmq2es : RabbitMq ReConnection Timeout ms( default 5000 )",
			Destination: &rmqReconnTimeout,
		},

		// Exchange CLI flags
		cli.StringFlag{
			Name:        "exname",
			Value:       "test",
			Usage:       "For rmq2es : Exchange name to declare ( Default test)",
			Destination: &exName,
		},
		cli.StringFlag{
			Name:        "exkind",
			Value:       "topic",
			Usage:       "For rmq2es : Exchange kind ( default topic)",
			Destination: &exKind,
		},

		// Rmq Queue CLI Flags
		cli.StringFlag{
			Name:        "qName",
			Value:       "test",
			Usage:       "For rmq2es : Queue name to declare ( Default test)",
			Destination: &qName,
		},
		cli.StringFlag{
			Name:        "qBindKey",
			Value:       "#",
			Usage:       "For rmq2es : Queue Binding Key with exchange ( default #)",
			Destination: &qBindKey,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "test",
			Aliases: []string{"th"},
			Usage:   "test",
			Action: func(c *cli.Context) error {

				// processRaw(`{ "delete": { "_index": "website", "_type": "blog", "_id": "123" }}
				//                { "create": { "_index": "website", "_type": "blog", "_id": "123" }}
				//                { "title":    "My first blog post" }
				//                { "index":  { "_index": "website", "_type": "blog" }}
				//                { "title":    "My second blog post" }
				//                { "update": { "_index": "website", "_type": "blog", "_id": "123", "_retry_on_conflict" : 3}}
				//                { "doc" : {"title" : "My updated blog post"}}
				//                `)

				dat, err := ioutil.ReadFile("./requests")
				if err != nil {
					panic(err)
				}
				fmt.Print(len(string(dat)))

				esBulkOps(dat)
				fmt.Printf("Done processing inputs")
				return nil
			},
		},
		{
			Name:    "threadpool",
			Aliases: []string{"th"},
			Usage:   "Show cluster threadpool",
			Action: func(c *cli.Context) error {

				return nil
			},
		},

		{
			Name:  "rmq2es",
			Usage: "RabbitMq to ES ingestion",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "dry-run",
					Usage: "dry run : Messge aren't inserted in ES, just printed on screen",
				},
			},
			Action: func(c *cli.Context) error {

				// check if Dry Run
				if !c.Bool("dry-run") {
					fmt.Println("Dry run flag disabled")
					dryrun = false
				} else {
					fmt.Println("Dry run flag enabled")
					dryrun = true
				}

				rmq2es()
				return nil
			},
		},
	}

	app.Run(os.Args)

}

/* Start process to consume data from Rmq and insert in ES */
func rmq2es() {

	initializeRmq()

	waitForever()
}

func waitForever() {
	// Waiting forever
	fmt.Printf("Waiting forever. Press Ctl+C to exit ...")
	forever := make(chan bool)
	<-forever
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
	var lType string
	var jumps int
	var iLines int = 0

	xulu.Use(splits, lType, jumps)

	// work on each split
	fmt.Printf("Splitting in lines : %d\n", len(splits))

	for iLines < len(splits) {
		fmt.Printf("Marshalling Line : %q\n", splits[iLines])
		parsedLine, is_sane := parseSplit(splits[iLines])

		if is_sane == true {
			lType, jumps = detectLineType(parsedLine)
			// fmt.Printf("%q\n", parsed.(type))

			iLines = iLines + jumps
		} else {
			iLines = iLines + 1
		}

	}

}

func parseSplit(singleLine string) (map[string]interface{}, bool) {

	// Json unmarshal
	// var stmnt1 esBulkStmntType
	// err := json.Unmarshal([]byte(splits[0]), &stmnt1)
	// if err != nil {
	//  // panic(err)
	//  log.Fatalf("json.Unmarshal: %s", err)
	// }

	// var parsed interface{}
	var parsed map[string]interface{}
	var is_sane bool = true

	err := json.Unmarshal([]byte(singleLine), &parsed)
	if err != nil {
		// panic(err)
		log.Printf("json.Unmarshal interface: %s", err)
		is_sane = false
	}

	// remarsh2, _ := json.Marshal(parsed)
	// xulu.Use(remarsh2)
	// map[delete:map[_id:123 _index:website _type:blog]]
	// fmt.Println(string(remarsh2))

	return parsed, is_sane

}

func detectLineType(unmarshalledLine map[string]interface{}) (string, int) {

	var jumps int
	var lType string

	// return line type and jump lines
	if unmarshalledLine["create"] != nil {
		lType = "create"

		// jump 2 lines
		jumps = 2

	} else if unmarshalledLine["delete"] != nil {
		lType = "delete"

		// jump 1 line
		jumps = 1

	} else if unmarshalledLine["index"] != nil {
		lType = "index"
		// jump 2 lines
		jumps = 2

	} else if unmarshalledLine["update"] != nil {
		lType = "update"

		// jump 2 lines
		jumps = 2

	} else {
		// No idea , simply skip this
		lType = "misc"

		// jump 1 line
		jumps = 1
	}

	fmt.Printf("Statement type : %q\n", lType)
	return lType, jumps
}

func parseThreadPoolOutpu(bulkData string) {
	/*
	   t := "id   pid   ip        host      bulk.active bulk.queue"
	   fmt.Printf("%q\n", strings.Fields(t))
	*/

}

func esGetThreadPool(bulkData string) {
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

func esBulkOps(bulkData []byte) {

	/*
		- Collect the message in buffer
		- See if Buffer needs to be Flushed
			- If yes, Flush
	*/

	/* Dry run */
	if dryrun == true {
		fmt.Println("Printing Dry Run Data")
		fmt.Println(string(bulkData))
		return
	}

	// Create bulk Uri
	url := esUrl + "/" + esIndex + "/_bulk"
	fmt.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(bulkData)))
	req.Header.Set("User-Agent", "elasty 1.0 - golang")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Es Bulk operation Error: %s", err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println("response Body:", string(body))
	xulu.Use(body)

}

// Re Initizlise rabbit Mq connection
func reInitializeRmq() {
	time.Sleep(time.Duration(rmqReconnTimeout) * time.Millisecond)
	initializeRmq()
}

// Initizlise rabbit Mq connection
func initializeRmq() {

	// Reconnect on conection Close
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Println("Reconnect after 5 seconds: " + err.Error())
		reInitializeRmq()
	}()

	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.DialConfig(rmqConnectStr, amqp.Config{FrameSize: 10240000})
	if err != nil {
		log.Println("Rmq Connection open: %s", err)
		reInitializeRmq()
	}
	conn.NotifyClose(c)
	fmt.Printf("Connection open\n")

	// Opening channel
	ch, err := conn.Channel()
	if err != nil {
		log.Println("Rmq Channel open: %s", err)
		reInitializeRmq()
	}
	ch.NotifyClose(c)
	fmt.Printf("Channel open\n")

	// Declare exchange
	err = ch.ExchangeDeclare(exName, exKind, true, false, false, false, nil)
	if err != nil {
		log.Println("Rmq Exchange Declare: %s", err)
		reInitializeRmq()
	}
	fmt.Printf("Exchange configured\n")

	// declare Queue
	q, err := ch.QueueDeclare(
		qName, // qname
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments table
	)
	if err != nil {
		log.Println("Rmq Q Declare: %s", err)
		reInitializeRmq()
	}
	fmt.Printf("Q configured\n")
	_ = q

	// Q bind
	err = ch.QueueBind(qName, qBindKey, exName, false, nil)
	if err != nil {
		log.Println("Rmq Q Bind: %s", err)
		reInitializeRmq()
	}
	fmt.Printf("Q bound\n")

	// Qos
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Println("Qos error: %s", err)
		reInitializeRmq()
	}

	//Setup consumer
	es_msgs, err := ch.Consume(qName, "go_consumer", false, false, false, false, nil)
	if err != nil {
		log.Println("Rmq Consumer Setup: %s", err)
		reInitializeRmq()
	}

	go func() {
		for each_msg := range es_msgs {
			// fmt.Printf("Msg: %s %s", string(each_msg.MessageId), string(each_msg.Body[:]))

			// send it to Elasticsearch as soon as you receive it .. and wait on receiving
			esBulkOps(each_msg.Body[:])
			err = each_msg.Ack(false)
			if err != nil {
				log.Fatalf("Error in Ack: %v", err)
			}
		}
	}()

}
