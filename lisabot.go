package main

import (
	"encoding/json"
	"errors"
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

var logger *lisaLog

func main() {
	confFile := flag.String("conf", "", "Conf files, you know, conf files")
	loglevel :=
		flag.String("loglevel", "warn", "log level: debug/info/warn/error")

	flag.Parse()

	var err error

	logger, err = NewLogger(os.Stdout, *loglevel)

	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		os.Exit(-1)
	}

	if *confFile == "" {
		logger.Error.Println("Need to specify a conf file")
		os.Exit(1)
	}

	confRaw, err := ioutil.ReadFile(*confFile)

	if err != nil {
		logger.Error.Println("Error reading config file: ", err)
		os.Exit(2)
	}

	var conf config
	err = yaml.Unmarshal(confRaw, &conf)

	if err != nil {
		logger.Error.Println("Error parsing config file: ", err)
	}

	if conf.Port == 0 {
		logger.Error.Println("No port specified!")
		os.Exit(3)
	}

	serverListener, err :=
		net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Ip, conf.Port))

	if err != nil {
		logger.Error.Println("Error opening socket for listening: ", err)
		os.Exit(4)
	}

	server, ok := serverListener.(*net.TCPListener)

	if !ok {
		logger.Error.Println("Listner isn't TCP? This is weird...")
		os.Exit(5)
	}

	quitChan := make(chan bool)

	dispatcherChan := make(chan *dispatcherRequest)

	go dispatcher(dispatcherChan, quitChan)

	logger.Info.Println("Server starting, entering main loop...")

	go listen(server, dispatcherChan)

	<-quitChan
	logger.Warn.Println("Termination requtested")

	logger.Warn.Println("Exited normally")
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
	if logger.Level == "debug" {
		debugReader, debugWriter := io.Pipe()
		streamIn = io.TeeReader(conn, debugWriter)
		go monitorRaw(debugReader)
	} else {
		streamIn = conn
	}

	decoder := json.NewDecoder(streamIn)
	encoder := json.NewEncoder(conn)

	var q *query
	id := ""
	for {
		q = new(query)
		err := decoder.Decode(q)

		if err != nil {
			logger.Error.Println(err)
			if err.Error() == "EOF" {
				break
			}
		} else {
			if id == "" {
				id, err = initialize(q, encoder, dispatcherChan)
				if err != nil {
					logger.Error.Println("Failed to engage:", err)
					conn.Close()
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
}

func initialize(q *query, encoder *json.Encoder,
	dispatcherChan chan *dispatcherRequest) (string, error) {

	if !q.validate() {
		return "", errors.New("Bad engagement request")
	}

	resp := make(chan string)

	dispatcherChan <- &dispatcherRequest{
		Query:      q,
		Encoder:    encoder,
		EngageResp: resp,
	}

	id := <-resp

	if id == "" {
		return "", errors.New("Error occured during engagement")
	}
	logger.Debug.Println("Connection successfully engaged")

	return id, nil
}

func monitorRaw(debugReader io.Reader) {
	buf := make([]byte, 2048)

	for {
		count, err := debugReader.Read(buf)

		if err == nil {
			logger.Debug.Println("Received: ", string(buf[:count]))
		} else {
			logger.Error.Println(err)
			if err.Error() == "EOF" {
				logger.Warn.Println("EOF detected")
				break
			}
		}
	}
}
