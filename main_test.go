package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strconv"
	"testing"
	"time"
)

func dummyWriter(text string, ch chan string) {
	for i := 0; i < 10; i++ {
		ch <- (text + " " + strconv.Itoa(i))
		time.Sleep(time.Millisecond * 200)
	}
}

func dummyReader(ch chan string) {
	for msg := range ch {
		fmt.Println(msg)
	}
}

func TestDummy(t *testing.T) {
	serverPublicKey, serverSecretKey, err := zmq.NewCurveKeypair()
	checkErr(err)
	client := NewClient("alice", "localhost", serverPublicKey)
	client2 := NewClient("bob", "localhost", serverPublicKey)
	server := NewServer(serverPublicKey, serverSecretKey)
	receiveChan := make(chan string, 1)
	receiveChan2 := make(chan string, 1)
	sendChan := make(chan string, 1)
	//sendChan2 := make(chan string, 1)
	go client.receiveMessages(receiveChan)
	go client2.receiveMessages(receiveChan2)
	go client.sendMessages(sendChan)
	go dummyWriter("hi all, i'm alice", sendChan)
	go dummyReader(receiveChan2)
	//go dummyChatter("that's great! i am charlie", client3)
	for i := 0; i < 10; i++ {
		message := server.getNextMessage()
		server.updateDisplays(message)
	}
}
