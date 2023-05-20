# go-iobox

Package implements inbox and outbox patterns for helping processing incoming and outgoing events.

## About

### Outbox

The command must atomically update the database and send messages in order to avoid data inconsistencies and bugs. 
However, it is not viable to use a traditional distributed transaction (2PC) that spans the database and the message 
broker The database and/or the message broker might not support 2PC. And even if they do,
it’s often undesirable to couple the service to both the database and the message broker.

But without using 2PC, sending a message in the middle of a transaction is not reliable.
There’s no guarantee that the transaction will commit. Similarly, if a service sends a message 
after committing the transaction there’s no guarantee that it won’t crash before sending the message.

The solution is for the service that sends the message to first store the message in the database as part of the 
transaction that updates the business entities. A separate process then sends the messages to the message broker.

[Source](https://microservices.io/patterns/data/transactional-outbox.html)

### Inbox

The inbox pattern is quite similar to the outbox pattern (but let’s say it works backwards). 
Then after receiving a new message, we don’t start processing right away, but only insert the message to 
the table and ACK. Finally, the background process picks up the rows from the inbox at a convenient 
pace and spins up processing. After the work is complete, the corresponding row in the table can 
be updated to mark the assignment as complete (or just removed from the inbox).

If received messages have any kind of unique key, they can be deduplicated before being saved to the inbox. 
Repeated messages can be caused by the crash of the recipient just after saving the row to the table, 
but before sending a successful ack.

[Source](https://softwaremill.com/microservices-101/#inbox-pattern)

## Install

To install the package you can use default Golang package installer.

```bash
go get github.com/Melenium2/go-iobox
```

## Example

### Outbox

Full example code you can saw [here](https://github.com/Melenium2/go-iobox/blob/master/example/outbox/publisher/main.go)

```go
func main() {
    broker := NewBroker()
    db := NewDbConn()

    // Create new outbox proccessor. Set the broker dependency to publish events 
    // and the database connection.
    ob := outbox.NewOutbox(broker, db)

    // Start the outbox processor. Function also intialize the outbox message table if it does not exists.
    if err = ob.Start(context.Background()); err != nil {
    	log.Fatal(err)
    }

    // ...

    // Outbox also provides the Client that saves outgoing events to the temporary table.
    // You should use it to push new messages to processor for future publishing.
    ouboxClient := ob.Writer()
}
```

### Inbox

...
