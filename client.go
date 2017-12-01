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

func (clnt *Client) keepAlive(ch chan []string) {
	for {
		message := &Message{
			Msg:  "",
			User: clnt.username,
			Type: 1,
		}

		message_json, err := json.Marshal(message)
		checkErr(err)
		clnt.chatSocket.Send(string(message_json), 0)
		// Wait for users list
		msgs, _ := clnt.chatSocket.RecvMessage(0)
		messager := &[]string{}
		json.Unmarshal([]byte(msgs[0]), messager)
		ch <- *messager
		time.Sleep(2 * time.Second)
	}
}

func (clnt *Client) receiveMessages(ch chan string) {
	for {
		message_string, err := clnt.displaySocket.Recv(0)
		checkErr(err)
		message := &Message{}
		json.Unmarshal([]byte(message_string), message)
		ch <- fmt.Sprintf("%v: %v\n", message.User, message.Msg)
	}
}

func (clnt *Client) sendMessages(ch chan string) {
	for message := range ch {
		clnt.sendMessage(message)
	}
}

func clientCommand(c *cli.Context) {
	serverPublicKey, err := readKeysFromFile(c.String("servercert-public"))
	checkErr(err)
	client := NewClient(c.String("username"), c.String("server-address"), serverPublicKey)

	receiveChan := make(chan string, 1)
	sendChan := make(chan string, 1)
	usersChan := make(chan []string, 1)
	go client.receiveMessages(receiveChan)
	go client.sendMessages(sendChan)
	go client.keepAlive(usersChan)
	showUI(receiveChan, sendChan, usersChan)
}
