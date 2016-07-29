package main

import (
    "fmt"
	"net/http"
//    "io/ioutil"
    "log"
    "strings"
    "flag"
)

func main() {

    // CLI
    bThreadpool := flag.Bool("threadpool", false, "Print threadpool info")
    flag.Parse()

    if bThreadpool {
        
        // call threadpool Uri
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
    }

    fmt.Printf("Body is, %q", body)
    */

    t := "id   pid   ip        host      bulk.active bulk.queue"

    fmt.Printf("%q\n", strings.Fields(t))
}

