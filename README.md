# elasty
elastic CLI for unix mac

# Config File
See `config/app.toml` , the config file has all the comments

# Build
Build assumes you have goimports in your system. If not install with `go get golang.org/x/tools/cmd/goimports`

use `make` to build for your system.

To test config file after the build
```
./_release/elasty -c ./config/app.toml chkconfig
```

# Feed Messages in Elasticsearch from RabbitMq
run `elasty rmq2es` to start rabbitMq consumer

## Bulk insert routine
Requests from RabbitMq are pulled and flushed when :
 - The max size of bulk query ( 10Mb default ) is reached either with a single message or multiple.
 - Max timeout is reached for a message. This will lead to a flush of whatever we have.
 - Max documents limit is reached : is calculated by number of rows / 2 taking an average that insert, update requests are 2 liners whereas delete is 1 liner. 

## Dry Run
Use -d / --dry-run flag to read messages from RabbitMq and print the HTTP statements which should be executed in ES
```
elasty rmq2es --dry-run
```

# Helpful Links
 - https://www.rabbitmq.com/tutorials/tutorial-one-go.html
 - https://godoc.org/github.com/streadway/amqp


# upstart script
Place the binary elasty_linux_amd64 in /root/elasty folder. And place the below script in /etc/init folder by the name of elasty.conf

```sh
description "elasty rabbitmq ingest Upstart script"
author "Shrey Agarwal"

start on (net-device-up
          and local-filesystems
          and runlevel [2345])
stop on runlevel [!2345]

env DAEMON=/usr/bin/elasty
env PID=/var/run/elasty.pid

respawn
respawn limit 5 100

script
    echo $$ > /var/run/elasty.pid
    exec $DAEMON rmq2es >> /var/log/elasty/elasty-service.log

end script

pre-start script
    echo "[`date`] Starting elasty" >> /var/log/elasty/elasty-service.log
end script

pre-stop script
    rm /var/run/elasty.pid
    echo "[`date`] Stopping elasty" >> /var/log/elasty/elasty-service.log
end script
```

#ToDo / Known Bugs

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

