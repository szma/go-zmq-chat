package main

import (
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"github.com/urfave/cli"
	"log"
	"time"
)

type Server struct {
	chatSocket    *zmq.Socket
	displaySocket *zmq.Socket
	usersLastSeen map[string]time.Time
}

func (srv *Server) updateUsers() []string {

	// Remove users that didnt call for 5 seconds
	now := time.Now()
	for k := range srv.usersLastSeen {
		if now.Sub(srv.usersLastSeen[k]) > 5*time.Second {
			delete(srv.usersLastSeen, k)
		}
	}

	// Create list of users from map
	keys := make([]string, 0, len(srv.usersLastSeen))
	for k := range srv.usersLastSeen {
		keys = append(keys, k)
	}

	return keys
}

func (srv *Server) checkNextMessage() {
	message_string, err := srv.chatSocket.RecvMessage(0)
	checkErr(err)
	identity := message_string[0]
	message := &Message{}
	json.Unmarshal([]byte(message_string[1]), message)

	srv.usersLastSeen[message.User] = time.Now()

	if message.Type == 0 {
		//Normal message
		log.Println(message_string)
		srv.updateDisplays(message)
		srv.chatSocket.SendMessage([]string{identity, "ok"}, 0)
	} else if message.Type == 1 {
		//Keep alive message: return users
		users := srv.updateUsers()
		message_json, _ := json.Marshal(users)
		srv.chatSocket.SendMessage([]string{identity, string(message_json)}, 0)
	}

}

func (srv *Server) updateDisplays(msg *Message) {
	message_json, err := json.Marshal(msg)
	checkErr(err)
	srv.displaySocket.Send(string(message_json), 0)
}

func NewServer(serverPublicKey, serverSecretKey string) *Server {
	server := &Server{}
	server.chatSocket, _ = zmq.NewSocket(zmq.ROUTER)
	server.displaySocket, _ = zmq.NewSocket(zmq.PUB)

	zmq.AuthCurveAdd("domain1", zmq.CURVE_ALLOW_ANY)
	server.chatSocket.ServerAuthCurve("domain1", serverSecretKey)
	server.displaySocket.ServerAuthCurve("domain1", serverSecretKey)

	err := server.chatSocket.Bind("tcp://*:5556")
	checkErr(err)
	err = server.displaySocket.Bind("tcp://*:5555")
	checkErr(err)

	server.usersLastSeen = make(map[string]time.Time)

	return server
}

func serverCommand(c *cli.Context) {
	if c.Bool("generate-certificate") {
		generateCertificates(c.String("servercert-public"), c.String("servercert-secret"))
	}
	serverPublicKey, err := readKeysFromFile(c.String("servercert-public"))
	serverSecretKey, err := readKeysFromFile(c.String("servercert-secret"))
	checkErr(err)

	server := NewServer(serverPublicKey, serverSecretKey)
	log.Println("Server started...")
	for {
		server.checkNextMessage()
	}
}
