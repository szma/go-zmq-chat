package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/urfave/cli"
	"log"
	"time"
)

type Client struct {
	chatSocket    *zmq.Socket
	displaySocket *zmq.Socket
	username      string

	receiveChan chan string
	sendChan    chan string
	usersChan   chan []string
}

func NewClient(username string, serverAddress string, serverPublicKey string) *Client {
	client := &Client{}

	client.chatSocket, _ = zmq.NewSocket(zmq.DEALER)
	client.displaySocket, _ = zmq.NewSocket(zmq.SUB)

	clientPublicKey, clientSecretKey, err := zmq.NewCurveKeypair()
	checkErr(err)
	client.chatSocket.ClientAuthCurve(serverPublicKey, clientPublicKey, clientSecretKey)
	client.displaySocket.ClientAuthCurve(serverPublicKey, clientPublicKey, clientSecretKey)

	client.chatSocket.SetIdentity(clientPublicKey)

	client.displaySocket.SetSubscribe("")
	err = client.chatSocket.Connect("tcp://" + serverAddress + ":5556")
	checkErr(err)
	err = client.displaySocket.Connect("tcp://" + serverAddress + ":5555")
	checkErr(err)
	client.username = username

	client.receiveChan = make(chan string, 1)
	client.sendChan = make(chan string, 1)
	client.usersChan = make(chan []string, 1)

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
	msgs, _ := clnt.chatSocket.RecvMessage(0)
	if msgs[0] != "ok" {
		log.Println("server response unexpected:", msgs)
	}
}

func (clnt *Client) keepAlive() {
	for {
		message := &Message{
			Msg:  "",
			User: clnt.username,
			Type: 1,
		}

		message_json, err := json.Marshal(message)
		checkErr(err)
		log.Println("Send ", clnt.username, string(message_json))
		clnt.chatSocket.Send(string(message_json), 0)
		// Wait for users list
		msgs, _ := clnt.chatSocket.RecvMessage(0)
		log.Println("Recv ", clnt.username, msgs)
		message_users := &[]string{}
		json.Unmarshal([]byte(msgs[0]), message_users)
		clnt.usersChan <- *message_users
		time.Sleep(2 * time.Second)
	}
}

func (clnt *Client) receiveMessages() {
	for {
		message_string, err := clnt.displaySocket.Recv(0)
		checkErr(err)
		message := &Message{}
		json.Unmarshal([]byte(message_string), message)
		clnt.receiveChan <- fmt.Sprintf("%v: %v\n", message.User, message.Msg)
	}
}

func (clnt *Client) sendMessages() {
	for message := range clnt.sendChan {
		clnt.sendMessage(message)
	}
}

func clientCommand(c *cli.Context) {
	serverPublicKey, err := readKeysFromFile(c.String("servercert-public"))
	checkErr(err)
	client := NewClient(c.String("username"), c.String("server-address"), serverPublicKey)

	go client.receiveMessages()
	go client.sendMessages()
	go client.keepAlive()
	showUI(client)
}
