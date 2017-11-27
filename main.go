package main

import (
	"bufio"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/urfave/cli"
	"log"
	"os"
	"runtime"
)

type Message struct {
	User string
	Msg  string
	Type int
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

func main() {
	zmq.AuthSetVerbose(false)
	zmq.AuthStart()
	app := cli.NewApp()
	app.Name = "go-zmq-chat"
	app.Usage = "Small chat program written in Go using ZeroMQ and encryption."
	app.Version = "0.1"
	app.Action = func(c *cli.Context) { fmt.Println("Start with client or server subcommand. --help for help.") }
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
