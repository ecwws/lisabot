package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/priscillachat/prislog"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type config struct {
	Port       int             `yaml:"port"`
	Ip         string          `yaml:"ip,omitempty"`
	Prefix     string          `yaml:"prefix"`
	Secret     string          `yaml:"secret"`
	Responders responderConfig `yaml:"responders"`
	prefixLen  int
}

type responderConfig struct {
	Passive []*passiveResponderConfig `yaml:"passive"`
}

type passiveResponderConfig struct {
	Match       string   `yaml:"match"`
	NoPrefix    bool     `yaml:"noprefix"`
	FallThrough bool     `yaml:"fallthrough"`
	Cmd         string   `yaml:"cmd"`
	Args        []string `yaml:"args"`
}

var logger *prislog.PrisLog
var conf config
var prefixPResponders []*passiveResponderConfig   // passive
var noPrefixPResponders []*passiveResponderConfig // passive

func main() {
	confFile := flag.String("conf", "", "Conf files, you know, conf files")
	loglevel :=
		flag.String("loglevel", "warn", "log level: debug/info/warn/error")

	flag.Parse()

	var err error

	logger, err = prislog.NewLogger(os.Stdout, *loglevel)

	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		os.Exit(1)
	}

	if *confFile == "" {
		logger.Error.Fatal("Need to specify a conf file")
	}

	confRaw, err := ioutil.ReadFile(*confFile)

	if err != nil {
		logger.Error.Fatal("Error reading config file: ", err)
	}

	err = yaml.Unmarshal(confRaw, &conf)

	if err != nil {
		logger.Error.Fatal("Error parsing config file: ", err)
	}

	logger.Debug.Println("Config loaded:", conf)

	prefixPResponders = make([]*passiveResponderConfig, 0)
	noPrefixPResponders = make([]*passiveResponderConfig, 0)

	for _, pr := range conf.Responders.Passive {
		logger.Debug.Println("Passive responder:", *pr)
		if pr.NoPrefix {
			noPrefixPResponders = append(noPrefixPResponders, pr)
		} else {
			prefixPResponders = append(prefixPResponders, pr)
		}
	}

	if conf.Port == 0 {
		logger.Error.Fatal("No port specified!")
	}

	// Prefix need to be free of excess spaces
	conf.Prefix = strings.Trim(conf.Prefix, " ")
	if len(conf.Prefix) < 1 {
		logger.Error.Fatal("Prefix is empty")
	}
	conf.Prefix += " "
	conf.prefixLen = len(conf.Prefix)

	serverListener, err :=
		net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Ip, conf.Port))

	if err != nil {
		logger.Error.Println("Error opening socket for listening: ", err)
		os.Exit(5)
	}

	server, ok := serverListener.(*net.TCPListener)

	if !ok {
		logger.Error.Println("Listner isn't TCP? This is weird...")
		os.Exit(6)
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
				dispatcherChan <- &dispatcherRequest{
					Query: &query{
						Type:   "command",
						Source: id,
						Command: &commandBlock{
							Action: "disengage",
						},
					},
				}
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
					// ignore the source identifier from the client, we'll
					// use the identifier returned from engagement
					q.Source = id
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
