package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"runtime"
	"strconv"
	//"time"
)

type Message struct {
	User string
	Msg  string
}

type Server struct {
	chatSocket    *zmq.Socket
	displaySocket *zmq.Socket
}

type Client struct {
	chatSocket    *zmq.Socket
	displaySocket *zmq.Socket
	username      string
}

func checkErr(err error) {
	if err != nil {
		log.SetFlags(0)
		_, filename, lineno, ok := runtime.Caller(1)
		if ok {
			log.Fatalf("%v:%v: %v", filename, lineno, err)
		} else {
			log.Fatalln(err)
		}
	}
}

func (srv *Server) getNextMessage() *Message {
	message_string, err := srv.chatSocket.Recv(0)
	checkErr(err)
	message := &Message{}
	json.Unmarshal([]byte(message_string), message)
	return message
}

func (srv *Server) updateDisplays(msg *Message) {
	message_json, err := json.Marshal(msg)
	checkErr(err)
	srv.displaySocket.Send(string(message_json), 0)
	srv.chatSocket.Send("ok", 0)
}

func NewServer() *Server {
	server := &Server{}
	server.chatSocket, _ = zmq.NewSocket(zmq.REP)
	server.displaySocket, _ = zmq.NewSocket(zmq.PUB)

	err := server.chatSocket.Bind("tcp://*:5556")
	checkErr(err)
	err = server.displaySocket.Bind("tcp://*:5555")
	checkErr(err)
	//defer server.chatSocket.Close()
	//defer server.displaySocket.Close()

	return server
}

func NewClient(username string, serverAddress string) *Client {
	client := &Client{}

	client.chatSocket, _ = zmq.NewSocket(zmq.REQ)
	client.displaySocket, _ = zmq.NewSocket(zmq.SUB)
	client.displaySocket.SetSubscribe("")

	err := client.chatSocket.Connect("tcp://" + serverAddress + ":5556")
	checkErr(err)
	err = client.displaySocket.Connect("tcp://" + serverAddress + ":5555")
	checkErr(err)
	client.username = username

	return client
}

func (clnt *Client) sendMessage(message_txt string) {
	message := &Message{
		Msg:  message_txt,
		User: clnt.username,
	}

	message_json, err := json.Marshal(message)
	checkErr(err)
	clnt.chatSocket.Send(string(message_json), 0)
	msgs, _ := clnt.chatSocket.Recv(0)
	if msgs != "ok" {
		log.Println("server response unexpected:", msgs)
	}
}

func (clnt *Client) receiveMessages() {
	for {
		message_string, err := clnt.displaySocket.Recv(0)
		checkErr(err)
		message := &Message{}
		json.Unmarshal([]byte(message_string), message)
		fmt.Printf("%v:\t%v\n", message.User, message.Msg)
	}
}

func dummyChatter(client *Client) {
	for i := 0; i < 10; i++ {
		client.sendMessage("Hi all" + strconv.Itoa(i))
	}
}

func main() {
	client := NewClient("alice", "localhost")
	server := NewServer()
	go dummyChatter(client)
	go client.receiveMessages()
	for {
		message := server.getNextMessage()
		server.updateDisplays(message)
	}
}
