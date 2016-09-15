package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lunux2008/xulu"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"github.com/urfave/cli"
)

// Log variables
var outlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

var dryrun bool

/* config maps */
var configInt = make(map[string]int)
var configBool = make(map[string]bool)
var configStr = make(map[string]string)

/* End global variables */

func main() {

	// Setting default variables before reading config file
	setDefaultConfigs()

	/* Read config file */
	readConfig()

	// setup logger
	redirectLogToFiles()

	/* signal handler */
	sigUSR1Handle()

	/* Main function only has CLI parsing */
	cliArgsParse()
}

/* Parse the CLI args and call appropriate function*/
func cliArgsParse() {

	app := cli.NewApp()
	app.Name = "elasty"
	app.Version = "0.0.3"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Shrey Agarwal",
			Email: "s@shreyagarwal.com",
		},
	}
	app.Usage = "Elasticsearch toolbelt based on experience"

	app.Flags = []cli.Flag{}

	app.Commands = []cli.Command{
		{
			Name:    "test",
			Aliases: []string{"th"},
			Usage:   "test",
			Action: func(c *cli.Context) error {

				dat, err := ioutil.ReadFile("./requests")
				if err != nil {
					panic(err)
				}
				outlog.Print(len(string(dat)))

				esBulkOps(dat)
				outlog.Printf("Done processing inputs")
				return nil
			},
		},
		{
			Name:  "threadpool",
			Usage: "Show cluster threadpool",
			Action: func(c *cli.Context) error {

				esGetThreadPool()
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
					outlog.Println("Dry run flag disabled")
					dryrun = false
				} else {
					outlog.Println("Dry run flag enabled")
					dryrun = true
				}

				rmq2es()
				return nil
			},
		},

		{
			Name:  "signal",
			Usage: "Send Signal to Pid",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				// syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
				return nil
			},
		},
	}

	app.Run(os.Args)

}

func setDefaultConfigs() {

	configStr["global.esUrl"] = "http://localhost:9200"

	configStr["global.access_log"] = ""
	configStr["global.error_log"] = ""

	configStr["rmq2es.rmqConnectString"] = "amqp://guest:guest@localhost:5672/"
	configInt["rmq2es.rmqReconnTimeout"] = 5000

	configBool["rmq2es.exDeclare"] = false
	configStr["rmq2es.exName"] = "test"
	configStr["rmq2es.exKind"] = "topic"
	configBool["rmq2es.exDurable"] = false
	configBool["rmq2es.exAutoDelete"] = false
	configBool["rmq2es.exInternal"] = false
	configBool["rmq2es.exNoWait"] = false

	configBool["rmq2es.qDeclare"] = false
	configStr["rmq2es.qName"] = "test"
	configBool["rmq2es.qDurable"] = true
	configBool["rmq2es.qAutoDelete"] = true
	configBool["rmq2es.qExclusive"] = false
	configBool["rmq2es.qNoWait"] = false

	configBool["rmq2es.qBind"] = false
	configStr["rmq2es.qBindKey"] = "#"
	configBool["rmq2es.qBindNoWait"] = false

	configStr["rmq2es.qConsumer"] = "elasty_consumer"
	configBool["rmq2es.qConsumeAutoAck"] = false
	configBool["rmq2es.qConsumeExclusive"] = false
	configBool["rmq2es.qConsumeNoLocal"] = false
	configBool["rmq2es.qConsumeNoWait"] = false

	configInt["rmq2es.prefetch_count"] = 1
	configInt["rmq2es.prefetch_size"] = 0
	configBool["rmq2es.prefetch_global"] = false

}

func readConfig() {

	viper.SetConfigName("app")    // no need to include file extension
	viper.AddConfigPath("config") // set the path of your config file

	err := viper.ReadInConfig()
	if err != nil {
		errlog.Println("Config file not found... at config/app.toml")
	} else {
		// outlog.Println("Reading Config File")

		// String configs
		for key := range configStr {
			if viper.IsSet(key) {
				configStr[key] = viper.GetString(key)
				// outlog.Println("Config: Key:", key, "Value:", configStr[key])
			} else {
				// outlog.Println("Default :", key, "Value:", value)
			}
		}

		// Int configs
		for key := range configInt {
			if viper.IsSet(key) {
				configInt[key] = viper.GetInt(key)
				// outlog.Println("Config: Key:", key, "Value:", configInt[key])
			} else {
				// outlog.Println("Default :", key, "Value:", value)
			}
		}
		// Bool configs
		for key := range configBool {
			if viper.IsSet(key) {
				configBool[key] = viper.GetBool(key)
				// outlog.Println("Config: Key:", key, "Value:", configBool[key])
			} else {
				// outlog.Println("Default :", key, "Value:", value)
			}
		}

	}

	// outlog.Println("Config Loaded\n")
}

/* redirect Logs to files */
func redirectLogToFiles() {

	// Setting Output Log
	if len(configStr["global.access_log"]) > 0 {

		fout, err := os.OpenFile(configStr["global.access_log"], os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			outlog.Printf("error opening out log file: %v", err)
		}

		err = syscall.Dup2(int(fout.Fd()), int(os.Stdout.Fd()))
		if err != nil {
			errlog.Fatalf("Failed to redirect stdout to file: %v", err)
		}
	}

	// Setting Error Log
	if len(configStr["global.error_log"]) > 0 {

		ferr, err := os.OpenFile(configStr["global.error_log"], os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			outlog.Printf("error opening error log file: %v", err)
		}

		err = syscall.Dup2(int(ferr.Fd()), int(os.Stderr.Fd()))
		if err != nil {
			errlog.Fatalf("Failed to redirect stderr to file: %v", err)
		}
	}

}

/* Singla handler to reset Log Files */
func sigUSR1Handle() {

	outlog.Printf("Signal handler set for Process %d \n\n", os.Getpid())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		sig := <-sigs
		outlog.Println()
		outlog.Println(sig)
	}()
}

/* Start process to consume data from Rmq and insert in ES */
func rmq2es() {

	initializeRmq()

	waitForever()
}

func waitForever() {
	// Waiting forever
	outlog.Printf("Waiting forever. Press Ctl+C to exit ...")
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
	// outlog.Printf("%d\n", len(strings.Split(rawData, "\n")))

	var splits []string = strings.Split(rawData, "\n")
	var lType string
	var jumps int
	var iLines int = 0

	xulu.Use(splits, lType, jumps)

	// work on each split
	outlog.Printf("Splitting in lines : %d\n", len(splits))

	for iLines < len(splits) {
		outlog.Printf("Marshalling Line : %q\n", splits[iLines])
		parsedLine, is_sane := parseSplit(splits[iLines])

		if is_sane == true {
			lType, jumps = detectLineType(parsedLine)
			// outlog.Printf("%q\n", parsed.(type))

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
	//  errlog.Fatalf("json.Unmarshal: %s", err)
	// }

	// var parsed interface{}
	var parsed map[string]interface{}
	var is_sane bool = true

	err := json.Unmarshal([]byte(singleLine), &parsed)
	if err != nil {
		// panic(err)
		errlog.Printf("json.Unmarshal interface: %s", err)
		is_sane = false
	}

	// remarsh2, _ := json.Marshal(parsed)
	// xulu.Use(remarsh2)
	// map[delete:map[_id:123 _index:website _type:blog]]
	// outlog.Println(string(remarsh2))

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

	outlog.Printf("Statement type : %q\n", lType)
	return lType, jumps
}

func parseThreadPoolOutput(bulkData string) {
	/*
	   t := "id   pid   ip        host      bulk.active bulk.queue"
	   outlog.Printf("%q\n", strings.Fields(t))
	*/

	// Split string by newlines
	for _, element := range strings.Split(bulkData, "\n") {
		if len(element) <= 0 {
			continue
		}
		outlog.Printf("%q\n", strings.Fields(element))
	}
}

func esGetThreadPool() {
	url := configStr["global.esUrl"] + "/_cat/thread_pool?v&h=id,pid,ip,host,bulk.active,bulk.queue"
	outlog.Println("URL:>", url)

	resp, err := http.Get(url)

	if err != nil {
		errlog.Fatalf("Threadpool HTTP Error Error: %s", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		errlog.Fatalf("Threadpool Http Response body, Error: %s", err)
	}

	parseThreadPoolOutput(string(body))
	// outlog.Printf("ThreadPool, %q", body)
}

func esBulkOps(bulkData []byte) {

	/*
		- Collect the message in buffer
		- See if Buffer needs to be Flushed
			- If yes, Flush
	*/

	/* Dry run */
	if dryrun == true {
		outlog.Println("Printing Dry Run Data")
		outlog.Println(string(bulkData))
		return
	}

	// Create bulk Uri
	url := configStr["global.esUrl"] + "/_bulk"
	outlog.Println("URL:>", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(bulkData)))
	req.Header.Set("User-Agent", "elasty 1.0 - golang")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errlog.Fatalf("Es Bulk operation Error: %s", err)
	}
	defer resp.Body.Close()

	outlog.Println("response Status:", resp.Status)
	outlog.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)

	// outlog.Println("response Body:", string(body))
	xulu.Use(body)

}

// Re Initizlise rabbit Mq connection
func reInitializeRmq() {
	time.Sleep(time.Duration(configInt["rmq2es.rmqReconnTimeout"]) * time.Millisecond)
	initializeRmq()
}

// Initizlise rabbit Mq connection
func initializeRmq() {

	// Reconnect on conection Close
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		errlog.Println("Reconnect after 5 seconds: " + err.Error())
		reInitializeRmq()
	}()

	// Connects opens an AMQP connection from the credentials in the URL.
	conn, err := amqp.DialConfig(configStr["rmq2es.rmqConnectString"], amqp.Config{FrameSize: 10240000})
	if err != nil {
		errlog.Println("Rmq Connection open: %s", err)
		reInitializeRmq()
	}
	conn.NotifyClose(c)
	outlog.Printf("Connection open\n")

	// Opening channel
	ch, err := conn.Channel()
	if err != nil {
		errlog.Println("Rmq Channel open: %s", err)
		reInitializeRmq()
	}
	ch.NotifyClose(c)
	outlog.Printf("Channel open\n")

	// Declare exchange
	if configBool["rmq2es.exDeclare"] {
		outlog.Printf("Declaring Exchange with settings name:%s exKind:%s exDurable:%t exAutoDelete:%t exInternal:%t exNoWait:%t\n",
			configStr["rmq2es.exName"],
			configStr["rmq2es.exKind"],
			configBool["rmq2es.exDurable"],
			configBool["rmq2es.exAutoDelete"],
			configBool["rmq2es.exInternal"],
			configBool["rmq2es.exNoWait"],
		)

		err = ch.ExchangeDeclare(
			configStr["rmq2es.exName"],
			configStr["rmq2es.exKind"],
			configBool["rmq2es.exDurable"],
			configBool["rmq2es.exAutoDelete"],
			configBool["rmq2es.exInternal"],
			configBool["rmq2es.exNoWait"],
			nil,
		)
		if err != nil {
			errlog.Println("Rmq Exchange Declare: %s", err)
			reInitializeRmq()
		}
		outlog.Printf("Exchange Declared\n\n")
	} else {
		outlog.Printf("Not declaring Exchange\n\n")
	}

	// declare Queue
	if configBool["rmq2es.qDeclare"] {
		outlog.Printf("Declaring Q with settings name:%s qDurable:%t qAutoDelete:%t qExclusive:%t qNoWait:%t\n",
			configStr["rmq2es.qName"],
			configBool["rmq2es.qDurable"],
			configBool["rmq2es.qAutoDelete"],
			configBool["rmq2es.qExclusive"],
			configBool["rmq2es.qNoWait"],
		)

		q, err := ch.QueueDeclare(
			configStr["rmq2es.qName"],        // qname
			configBool["rmq2es.qDurable"],    // durable
			configBool["rmq2es.qAutoDelete"], // delete when unused
			configBool["rmq2es.qExclusive"],  // exclusive
			configBool["rmq2es.qNoWait"],     // no-wait
			nil, // arguments table
		)
		if err != nil {
			errlog.Println("Rmq Q Declare: %s", err)
			reInitializeRmq()
		}
		outlog.Printf("Q Declared\n\n")
		_ = q
	} else {
		outlog.Printf("Not declaring Queue\n\n")
	}

	// Q bind
	if configBool["rmq2es.qBind"] {
		err = ch.QueueBind(
			configStr["rmq2es.qName"],
			configStr["rmq2es.qBindKey"],
			configStr["rmq2es.exName"],
			configBool["rmq2es.qBindNoWait"],
			nil,
		)
		if err != nil {
			errlog.Println("Rmq Q Bind: %s", err)
			reInitializeRmq()
		}
		outlog.Printf("Q bound\n")
	} else {
		outlog.Printf("Not Binding Queue")
	}

	// Qos
	err = ch.Qos(
		configInt["rmq2es.prefetch_count"],
		configInt["rmq2es.prefetch_size"],
		configBool["rmq2es.prefetch_global"],
	)
	if err != nil {
		errlog.Println("Qos error: %s", err)
		reInitializeRmq()
	}

	//Setup consumer ... queue, consumer string, autoAck, exclusive, noLocal, noWait
	es_msgs, err := ch.Consume(
		configStr["rmq2es.qName"],
		configStr["rmq2es.qConsumer"],
		configBool["rmq2es.qConsumeAutoAck"],
		configBool["rmq2es.qConsumeExclusive"],
		configBool["rmq2es.qConsumeNoLocal"],
		configBool["rmq2es.qConsumeNoWait"],
		nil,
	)
	if err != nil {
		errlog.Println("Rmq Consumer Setup: %s", err)
		reInitializeRmq()
	}

	go func() {
		for each_msg := range es_msgs {
			// outlog.Printf("Msg: %s %s", string(each_msg.MessageId), string(each_msg.Body[:]))

			// send it to Elasticsearch as soon as you receive it .. and wait on receiving
			esBulkOps(each_msg.Body[:])
			err = each_msg.Ack(false)
			if err != nil {
				errlog.Fatalf("Error in Ack: %v", err)
			}
		}
	}()

}
