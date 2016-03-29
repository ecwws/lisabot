package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
)

type config struct {
	Port   int    `yaml:"port"`
	Ip     string `yaml:"ip,omitempty"`
	Secret string `yaml:"secret"`
}

type commandBlock struct {
	Source string `json:"sender"`
}

type messageBlock struct {
}

type block struct {
	BlockType string        `json:"type"`
	Source    string        `json:"source"`
	Command   *commandBlock `json:"command"`
	Message   *messageBlock `json:"message"`
}

func main() {
	confFile := flag.String("conf", "", "Conf files, you know, conf files")

	flag.Parse()

	if *confFile == "" {
		fmt.Println("Need to specify a conf file")
		os.Exit(1)
	}

	confRaw, err := ioutil.ReadFile(*confFile)

	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(2)
	}

	var conf config
	err = yaml.Unmarshal(confRaw, &conf)

	if err != nil {
		fmt.Println("Error parsing config file: ", err)
	}

	if conf.Port == 0 {
		fmt.Println("No port specified!")
		os.Exit(3)
	}

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Ip, conf.Port))

	if err != nil {
		fmt.Println("Error opening socket for listening: ", err)
		os.Exit(4)
	}

	quitChan := make(chan int)
	connChan := make(chan *net.Conn)

	fmt.Println("Server starting, entering main loop...")

	go listen(server, connChan)

MainLoop:
	for {
		select {
		case conn := <-connChan:
			fmt.Println("Connection accepted!")
			go serve(conn, quitChan)
		case <-quitChan:
			fmt.Println("Termination requtested")
			break MainLoop
		}
	}

	fmt.Println("Exited normally")
}

func listen(server net.Listener, connChan chan *net.Conn) {
	for {
		conn, err := server.Accept()
		if err == nil {
			connChan <- &conn
		}
	}
}

func serve(conn *net.Conn, quitChan chan int) {
	buf := make([]byte, 128)

	for {
		count, err := (*conn).Read(buf)

		if err != nil {
			fmt.Println("Error: ", err)
			if err.Error() == "EOF" {
				fmt.Println("EOF detected")
				break
			}
		}

		str := string(buf[:count])

		fmt.Println("Received: ", count)

		if str == "quit\n" || str == "quit\r\n" {
			fmt.Println("Quit command received")
			quitChan <- 1
		}
	}
}
