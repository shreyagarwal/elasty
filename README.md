# elasty
elastic CLI for unix mac

# Config File
See `config/app.toml` , the config file has all the comments


# Feed Messages in Elasticsearch from RabbitMq

## Bulk insert routine
Requests from RabbitMq are pulled and flusehed when :
 - The max size of bulk query ( 10Mb default ) is reached either with a single message or multiple.
 - Max timeout is reached for a message. This will lead to a flush of whatever we have.
 - Max documents limit is reached : is calculated by number of rows / 2 taking an average that insert, update requests are 2 liners whereas delete is 1 liner. 

## Dry Run
Use -d / --dry-run flag to read messages from RabbitMq and print the HTTP statements which should be executed in ES

### Algo
- Lets not Ack unless the message has been sent to ES.
- Prefetch increase or decrease will only throttle client which is not important here.

- At each message, count number of operations. There can be 4 types of operations : Create, index, update, delete. Parse each line, and see what the key is . Count operations accordingly.
- check
    - If total unacked message buffer > the buffer to flush ... normally keep this as prefetch_size only
    - If total number of unacked message > the messages to flush ... Normally keep this as PREFETCH COUNT only

=> Apart from it , if x ms have elapsed after oldest message , then flush it please ... setting timer per second isnot a nice idea.. 
A setting should say how long can a message survive if not flushed. Lets keep a timer HALF of that time...


# Helpful Links
 - https://www.rabbitmq.com/tutorials/tutorial-one-go.html
 - https://godoc.org/github.com/streadway/amqp

#ToDo / Expectations

 - CLI
     + url to support multiple urls of ES cluster later
     + Threadpool check. Poper format
     + Single Level check on ES , as what all the problems can be
     + Ncurses type tool to give all info
     + replica change of index
     + shard allocation ON/off
     + term tool to edit Config ... Options and values
     + Query Help : Hits, time taken, etc.

 - Insert bulk data in ES
     + Error handling : In case of ANY ES error do not HIT Nack, but fail that message Or retry x times, put in the end of Queue yourself after some time. Policy to be framed and decided
     + Buffer Data before pushing : Flush the data and Ack the messages
        + Flush Data based on document count
        + Flush Data in Size
        + Flush Data based on Time
     + Schema chek in case of bulk insert. helps in indentifying errors
     + Force Flush on signal receive
     + Check Threadpool before eash insert if all is OK 

 - Distribution
     + Create brew
     + Daemonize and Deb package
         * Write logs to files
         * Handle USR1 to re-open logs
         * Have nginx type start-stop daemon and signal handling CLI

