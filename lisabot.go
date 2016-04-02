package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
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
	Id      string   `json:"id"`
	Action  string   `json:"action"`
	Type    string   `json:"type"`
	Time    int      `json:"time"`
	Data    string   `json:"data"`
	Array   []string `json:"array"`
	Options []string `json:"options"`
}

type messageBlock struct {
	Message string `json:"message"`
	From    string `json:"from"`
	Room    string `json:"room"`
}

type query struct {
	Type    string        `json:"type"`
	Source  string        `json:"source"`
	Command *commandBlock `json:"command"`
	Message *messageBlock `json:"message"`
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

	debugReader, debugWriter := io.Pipe()
	streamIn := io.TeeReader(*conn, debugWriter)

	go monitor(debugReader)

	decoder := json.NewDecoder(streamIn)

	var block query
	for {
		err := decoder.Decode(&block)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		fmt.Println("Type: ", block.Type)
		fmt.Println("Source: ", block.Source)
	}
}

func monitor(debugReader io.Reader) {
	buf := make([]byte, 2048)

	for {
		count, err := debugReader.Read(buf)

		if err == nil {
			fmt.Println("Debug: ", string(buf[:count]))
		} else {
			fmt.Println("Error: ", err)
			if err.Error() == "EOF" {
				fmt.Println("EOF detected")
				break
			}
		}
	}
}
