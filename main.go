package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	kafka "github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
)

var (
	createRoom = flag.String("c", "_", "create new room")
	addr       = flag.String("addr", ":8080", "http service address")

	err error
)

var (
	//kafka
	topicKey  = "chat"
	kafkaHost = "localhost:9092"
	producer  kafka.AsyncProducer

	roomQ    = make(map[int]kafka.PartitionConsumer)
	roomList = []uint32{}
	hub      = newHub()
)

var (
	//mysql
	mysqlHost = "root:root@tcp(localhost:3306)/feed"
	mysqlDB   *sql.DB
)

var (
	u = user{userID: "chat_room"}
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func wsServer() {
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	flag.Parse()
	go wsServer()

	//init mysql con
	mysqlDB, err = sql.Open("mysql", mysqlHost)
	if err != nil {
		panic(err)
	}

	err = mysqlDB.Ping()
	if err != nil {
		panic(err)
	}

	//init kafka producer
	producer, err = kafka.NewAsyncProducer([]string{kafkaHost}, nil)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	//init kafka consumer
	consumer, err := kafka.NewConsumer([]string{kafkaHost}, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partitions, _ := consumer.Partitions(topicKey)
	for roomID := range partitions {
		partitionConsume, err := consumer.ConsumePartition(topicKey, int32(roomID), kafka.OffsetNewest)
		if err != nil {
			panic(err)
		}
		if roomQ[roomID] == nil {
			roomQ[roomID] = partitionConsume
		}
	}

	// partitionConsume, err := consumer.ConsumePartition("test", 0, kafka.OffsetNewest)
	// if err != nil {
	// 	panic(err)
	// }

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	ticker := time.NewTicker(1 * time.Second)

loop:
	for {
		select {
		case <-ticker.C:
			mysqlDB.Ping()
		case msg := <-roomQ[0].Messages():
			u.Post(string(msg.Value))
			hub.broadcast <- msg.Value
		case <-signals:
			break loop
		}
	}

	return
}
