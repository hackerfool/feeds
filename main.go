package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

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
	g = gin.Default()
)

func httpServer() {
	{
		go hub.run()
		// g.StaticFile("/", "web/home.html")
		g.StaticFile("/", "web/index.html")
		g.StaticFile("/web/login.html", "web/login.html")
		g.GET("/ws", serveWs)
	}

	v1 := g.Group("/v1")
	{
		v1.GET("/sign", userSign)
		v1.GET("/login", userLogin)
		v1.GET("/follow", vaildLogin, userFollow)
		v1.GET("/fans", vaildLogin, userFans)
		v1.POST("/post", vaildLogin, userPost)
		v1.GET("/postlist", vaildLogin, userPostList)
	}

	err := g.Run(*addr)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	go httpServer()

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
		// case msg := <-roomQ[0].Messages():
		// 	hub.broadcast <- msg.Value
		case <-signals:
			break loop
		}
	}

	return
}
