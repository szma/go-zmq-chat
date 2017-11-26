package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/urfave/cli"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
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

func NewServer(serverPublicKey, serverSecretKey string) *Server {
	server := &Server{}
	server.chatSocket, _ = zmq.NewSocket(zmq.REP)
	server.displaySocket, _ = zmq.NewSocket(zmq.PUB)

	zmq.AuthCurveAdd("domain1", zmq.CURVE_ALLOW_ANY)
	server.chatSocket.ServerAuthCurve("domain1", serverSecretKey)
	server.displaySocket.ServerAuthCurve("domain1", serverSecretKey)

	err := server.chatSocket.Bind("tcp://*:5556")
	checkErr(err)
	err = server.displaySocket.Bind("tcp://*:5555")
	checkErr(err)
	//defer server.chatSocket.Close()
	//defer server.displaySocket.Close()

	return server
}

func NewClient(username string, serverAddress string, serverPublicKey string) *Client {
	client := &Client{}

	client.chatSocket, _ = zmq.NewSocket(zmq.REQ)
	client.displaySocket, _ = zmq.NewSocket(zmq.SUB)

	clientPublicKey, clientSecretKey, err := zmq.NewCurveKeypair()
	checkErr(err)
	client.chatSocket.ClientAuthCurve(serverPublicKey, clientPublicKey, clientSecretKey)
	client.displaySocket.ClientAuthCurve(serverPublicKey, clientPublicKey, clientSecretKey)

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
		if message.User != clnt.username {
			fmt.Printf("%v:\t%v\n", message.User, message.Msg)
		}
	}
}

func dummyChatter(text string, client *Client) {
	for i := 0; i < 10; i++ {
		client.sendMessage(text + " " + strconv.Itoa(i))
		time.Sleep(time.Millisecond * 200)
	}
}

func dummyTest(c *cli.Context) {
	serverPublicKey, serverSecretKey, err := zmq.NewCurveKeypair()
	checkErr(err)
	client := NewClient("alice", "localhost", serverPublicKey)
	client2 := NewClient("bob", "localhost", serverPublicKey)
	client3 := NewClient("charlie", "localhost", serverPublicKey)
	server := NewServer(serverPublicKey, serverSecretKey)
	go client.receiveMessages()
	go client2.receiveMessages()
	go client3.receiveMessages()
	go dummyChatter("hi all", client)
	go dummyChatter("that's great!", client3)
	for {
		message := server.getNextMessage()
		server.updateDisplays(message)
	}
}

func readKeysFromFile(keyfile string) (key string, err error) {
	file, err := os.Open(keyfile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	key = scanner.Text()

	if err = scanner.Err(); err != nil {
		return
	}

	return
}

func generateCertificates(filenamePub, filenameSecret string) {
	publicKey, secretKey, err := zmq.NewCurveKeypair()
	checkErr(err)

	f, err := os.Create(filenamePub)
	checkErr(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "%v\n", publicKey)
	checkErr(err)
	w.Flush()

	f, err = os.Create(filenameSecret)
	checkErr(err)
	defer f.Close()
	w = bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "%v\n", secretKey)
	checkErr(err)
	w.Flush()
}

func clientCommand(c *cli.Context) {
	serverPublicKey, err := readKeysFromFile("server_cert.pub")
	checkErr(err)
	client := NewClient(c.String("username"), c.String("server-address"), serverPublicKey)
	go client.receiveMessages()
	dummyChatter("hi all", client)
}

func serverCommand(c *cli.Context) {
	if c.Bool("generate-certificate") {
		generateCertificates("server_cert.pub", "server_cert")
	}
	serverSecretKey, err := readKeysFromFile("server_cert")
	serverPublicKey, err := readKeysFromFile("server_cert.pub")
	checkErr(err)

	server := NewServer(serverPublicKey, serverSecretKey)
	log.Println("Server started...")
	for {
		message := server.getNextMessage()
		log.Println(message)
		server.updateDisplays(message)
	}
}

func main() {
	zmq.AuthSetVerbose(true)
	zmq.AuthStart()
	app := cli.NewApp()
	app.Name = "go-zmq-chat"
	app.Usage = "Small chat program written in Go using ZeroMQ and encryption."
	app.Version = "0.1"
	app.Action = dummyTest
	app.Commands = []cli.Command{
		{
			Name:   "server",
			Usage:  "run a chat server",
			Action: serverCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "generate-certificate, g",
					Usage: "Generate certificate files.",
				},
			},
		},
		{
			Name:   "client",
			Usage:  "run a chat client",
			Action: clientCommand,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server-address, url, s",
					Usage: "Server address",
					Value: "localhost",
				},
				cli.StringFlag{
					Name:  "username, u",
					Usage: "User name",
					Value: "guest",
				},
			},
		},
	}
	app.Run(os.Args)

	zmq.AuthStop()
}
