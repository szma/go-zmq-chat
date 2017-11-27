go-zmq-chat
=========

This is a simple chat tool to get into the Go programming language while also learning something about the ZeroMQ security features.
Kudos to https://github.com/jnthnwn/zmq-chat for this simple python version.

This is under development and not finished, yet. That means you cannot chat right now, but there are some dummy clients.
Next:
  * [x] Add encryption 
  * [x] Split client and server commnd
  * [x] Add simple user interface
  * [x] Send/show user list
  * [ ] Allow users to register names.


Disclaimer
----------

This is one of my first Go programs. Feel free to contribute or send me some hints to improve the program.

Installation
------------

```bash
$ go get github.com/szma/go-zmq-chat
```

If the $GOPATH/bin folder is in your $PATH run


```bash
$ go-zmq-chat server # add -g to generate new certificates.
$ # clients need to have the correct server_cert.pub!
$ go-zmq-chat client --username mike --url 192.168.12.23 # IP of your server
```


