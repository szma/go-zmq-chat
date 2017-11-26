package main

import (
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"github.com/urfave/cli"
	"log"
)

type Server struct {
	chatSocket    *zmq.Socket
	displaySocket *zmq.Socket
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
