package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type config struct {
	Port   int    `yaml:"port"`
	Ip     string `yaml:"ip,omitempty"`
	Secret string `yaml:"secret"`
}

var debugOut bool
var logerr *log.Logger
var logstd *log.Logger

func main() {
	confFile := flag.String("conf", "", "Conf files, you know, conf files")
	flag.BoolVar(&debugOut, "debug", false, "Debug output (true or false)")

	flag.Parse()

	logerr = log.New(os.Stderr, "Lisa Error", log.LstdFlags|log.Lshortfile)
	logstd = log.New(os.Stdout, "Lisa Out", log.LstdFlags|log.Lshortfile)

	if *confFile == "" {
		logerr.Println("Need to specify a conf file")
		os.Exit(1)
	}

	confRaw, err := ioutil.ReadFile(*confFile)

	if err != nil {
		logerr.Println("Error reading config file: ", err)
		os.Exit(2)
	}

	var conf config
	err = yaml.Unmarshal(confRaw, &conf)

	if err != nil {
		logerr.Println("Error parsing config file: ", err)
	}

	if conf.Port == 0 {
		logerr.Println("No port specified!")
		os.Exit(3)
	}

	serverListener, err :=
		net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Ip, conf.Port))

	if err != nil {
		logerr.Println("Error opening socket for listening: ", err)
		os.Exit(4)
	}

	server, ok := serverListener.(*net.TCPListener)

	if !ok {
		logerr.Println("Listner isn't TCP? This is weird...")
		os.Exit(5)
	}

	quitChan := make(chan int)

	dispatcherChan := make(chan *dispatcherRequest)

	go dispatcher(dispatcherChan, quitChan)

	logstd.Println("Server starting, entering main loop...")

	go listen(server, dispatcherChan)

	<-quitChan
	logstd.Println("Termination requtested")

	logstd.Println("Exited normally")
}

func listen(server *net.TCPListener, dispatcherChan chan *dispatcherRequest) {

	for {
		conn, err := server.AcceptTCP()
		if err == nil {
			go serve(conn, dispatcherChan)
		}
	}
}

func serve(conn *net.TCPConn, dispatcherChan chan *dispatcherRequest) {

	var streamIn io.Reader
	if debugOut {
		debugReader, debugWriter := io.Pipe()
		streamIn = io.TeeReader(conn, debugWriter)
		go monitorRaw(debugReader)
	} else {
		streamIn = conn
	}

	decoder := json.NewDecoder(streamIn)
	encoder := json.NewEncoder(conn)

	var q *query
	for {
		q = new(query)
		err := decoder.Decode(q)

		if err != nil {
			logerr.Println("Error: ", err)
			if err.Error() == "EOF" {
				break
			}
		} else {
			if q.validate() {
				dispatcherChan <- &dispatcherRequest{
					Query:   q,
					Encoder: encoder,
				}
			}
		}
	}
}

func monitorRaw(debugReader io.Reader) {
	buf := make([]byte, 2048)

	for {
		count, err := debugReader.Read(buf)

		if err == nil {
			logstd.Println("Debug: ", string(buf[:count]))
		} else {
			logerr.Println("Error: ", err)
			if err.Error() == "EOF" {
				logerr.Println("EOF detected")
				break
			}
		}
	}
}
