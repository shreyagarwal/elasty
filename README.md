# elasty
elastic CLI for unix mac


# Feed Messages in Elasticsearch from RabbitMq

# Config object
 - **uri** : Can be single value like `http://localhost:9200` , `https://localhost:9200` or can be an array to indicate cluster `['http://10.0.0.1:9200', 'http://10.0.0.2:9200'] . NOTE: Please specify full uri, i.e. with protocol and port.
 - **bulkQueueThreshold** : While doing bulk operation in ES, Threshold count is checked before each bulk op. If value returned for any of the nodes is > than this value, then operation wont be done on that node, and another node will be tried. Default : 2
 - **bulkThresholdRetries** : If bulk op request fails repeatedly on each node, then after max tries, the operation is delared failed. 0 or negative value indicates retry forever. Default : no. of hosts * bulkQueueThreshold * 2 . 
 - **bulkThresholdRetryInterval** : Retry interval for bulk operations, if it fails due to whatever reason, value in Milliseconds. Default : 2000 

# Bulk insert routine
Requests from RabbitMq are pulled and flusehed when :
 - The max size of bulk query ( 10Mb default ) is reached either with a single message or multiple.
 - Max timeout is reached for a message. This will lead to a flush of whatever we have.
 - Max documents limit is reached : is calculated by number of rows / 2 taking an average that insert, update requests are 2 liners whereas delete is 1 liner. 

### Algo
- Lets not Ack unless the message has been sent to ES.
- Prefetch increase or decrease will only throttle client which is not important here.

- At each message, count number of operations. There can be 4 types of operations : Create, index, update, delete. Parse each line, and see what the key is . Count operations accordingly.
- check
    - If total unacked message buffer > the buffer to flush ... normally keep this as prefetch_size only
    - If total number of unacked message > the messages to flush ... Normally keep this as PREFETCH COUNT only

=> Apart from it , if x ms have elapsed after oldest message , then flush it please ... setting timer per second isnot a nice idea.. 
A setting should say how long can a message survive if not flushed. Lets keep a timer HALF of that time...



#ToDo / Expectations

 - CLI
     + Threadpool check
     + Single Level check on ES , as what all the problems can be
     + Ncurses type tool to give all info
     + replica change of index
     + shard allocation ON/off
     + term tool to edit Config ... Options and values
     + Query Help : Hits, time taken, etc.

 - Insert bulk data in ES
     + Prog should be blocking without Sleep
     - Flush Data based on document count
     - Flush Data based on Time
     - Schema chek in case of bulk insert. helps in indentifying errors
     - Force Flush on signal receive
     - Check Threadpool before eash insert if all is OK 

 - Create brew , Deb file service


