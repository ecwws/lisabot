package main

import (
	"container/list"
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
	"regexp"
	"strings"
)

type config struct {
	Port       int              `yaml:"port"`
	Ip         string           `yaml:"ip,omitempty"`
	Prefix     string           `yaml:"prefix"`
	PrefixAlt  []string         `yaml:"prefix-alit"`
	Help       string           `yaml:"help-command"`
	Secret     string           `yaml:"secret"`
	Responders *responderConfig `yaml:"responders"`
	prefixLen  int
	helpRegex  *regexp.Regexp
}

type responderConfig struct {
	Passive []*passiveResponderConfig `yaml:"passive"`
}

type passiveResponderConfig struct {
	Name            string   `yaml:"name"`
	Match           []string `yaml:"match"`
	MentionMatch    []string `yaml:"mentionmatch"`
	NoPrefix        bool     `yaml:"noprefix"`
	FallThrough     bool     `yaml:"fallthrough"`
	Cmd             string   `yaml:"cmd"`
	Args            []string `yaml:"args"`
	Help            string   `yaml:"help"`
	HelpCmds        []string `yaml:"help-commands"`
	HelpMentionCmds []string `yaml:"help-mention-commands"`
	regex           []*regexp.Regexp
	mRegex          []*regexp.Regexp
	substitute      map[int]bool
}

type activeResponderConfig struct {
	regex     *regexp.Regexp
	source    string
	id        string
	matchNext bool
	helpCmd   string
	help      string
}

type helpInfo struct {
	helpCmd  string
	helpMsg  string
	noPrefix bool
	mention  bool
}

var logger *prislog.PrisLog
var conf config
var prefixPResponders *list.List
var noPrefixPResponders *list.List
var mentionPResponders *list.List
var unhandledPResponders *list.List

var prefixAResponders *list.List
var noPrefixAResponders *list.List
var mentionAResponders *list.List
var unhandledAResponders *list.List

var subRegex *regexp.Regexp
var help *list.List

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

	if conf.Help == "" {
		conf.Help = "help"
	}

	conf.helpRegex, err = regexp.Compile("^" + conf.Help + "\\s*(\\w)*")

	if err != nil {
		logger.Error.Fatal("Bad help command:", err)
	}

	logger.Debug.Println("Help command:", conf.helpRegex)

	logger.Debug.Println("Config loaded:", conf)

	prefixPResponders = list.New()
	noPrefixPResponders = list.New()
	mentionPResponders = list.New()

	prefixAResponders = list.New()
	noPrefixAResponders = list.New()
	mentionAResponders = list.New()
	unhandledAResponders = list.New()

	subRegex = regexp.MustCompile("__([[:digit:]])__")

	help = list.New()

	for _, pr := range conf.Responders.Passive {
		logger.Debug.Println("Passive responder:", *pr)

		if len(pr.Match) == 0 {
			logger.Error.Fatal(
				"Must specify at least one match for passive responder")
		}

		pr.regex = make([]*regexp.Regexp, 0)
		for _, pattern := range pr.Match {
			rg, err := regexp.Compile(pattern)
			if err != nil {
				logger.Error.Fatal("Unable to parse expression:", pattern)
			}
			pr.regex = append(pr.regex, rg)
		}

		pr.mRegex = make([]*regexp.Regexp, 0)
		for _, pattern := range pr.MentionMatch {
			rg, err := regexp.Compile(pattern)
			if err != nil {
				logger.Error.Fatal("Unable to parse expression:", pattern)
			}
			pr.mRegex = append(pr.mRegex, rg)
		}

		if len(pr.regex) == 0 {
			logger.Error.Fatal("Missing match or multimatch:", pr.Name)
		}

		if pr.Cmd == "" {
			logger.Error.Fatal(
				"Passive Responder must have 'cmd' paramenter")
		}

		pr.substitute = make(map[int]bool)
		for i, arg := range pr.Args {
			if ms := subRegex.MatchString(arg); ms {
				pr.substitute[i] = true
			}
		}

		if pr.NoPrefix {
			logger.Debug.Println("Registered NoPrefix responder:", pr.Name)
			noPrefixPResponders.PushBack(pr)
		} else {
			logger.Debug.Println("Registered Prefix responder:", pr.Name)
			prefixPResponders.PushBack(pr)
		}

		if len(pr.mRegex) != 0 {
			logger.Debug.Println("Registered Mention responder:", pr.Name)
			mentionPResponders.PushBack(pr)
		}

		if pr.Help == "" || len(pr.HelpCmds) == 0 {
			logger.Error.Fatal(
				"Missing help or help-commands for passive responder: ",
				pr.Name)
		}

		for _, cmd := range pr.HelpCmds {
			info := &helpInfo{
				helpCmd: cmd,
				helpMsg: pr.Help,
			}

			if pr.NoPrefix {
				info.noPrefix = true
			}

			help.PushBack(info)
		}

		for _, cmd := range pr.HelpMentionCmds {
			help.PushBack(&helpInfo{
				helpCmd: cmd,
				helpMsg: pr.Help,
				mention: true,
			})
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
	isAdapter := false
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

				if q.Command.Type == "adapter" {
					isAdapter = true
				}
			} else {
				if err := q.validate(); err == nil {
					// ignore the source identifier from the client, we'll
					// use the identifier assigned during engagement
					q.Source = id

					// if message is from adapter, ignore the value of the "to"
					// field, it should always be empty or "server"
					if isAdapter {
						// only info reply allowed to pass directly from adapter
						// to responder
						if q.Type != "command" || q.Command.Action != "info" {
							q.To = ""
						}

						if q.Type == "command" &&
							q.Command.Action == "register" {

							logger.Error.Println(
								"Adapter cannot register commands")
							continue
						}
					} else if q.To == "" {
						// don't forward the responder message that is missing
						// "to" field, this could potentially cause an infinite
						// loop
						logger.Error.Println(
							"Responder query missing 'to' field")
						continue
					} else if q.Type == "message" && q.To == "server" {
						logger.Error.Println(
							"Responder message cannot target 'server'")
						continue
					}
					dispatcherChan <- &dispatcherRequest{
						Query:   q,
						Encoder: encoder,
					}
				} else {
					logger.Error.Println("Failed to validate query:", err)
				}
			}
		}
	}
}

func initialize(q *query, encoder *json.Encoder,
	dispatcherChan chan *dispatcherRequest) (string, error) {

	if err := q.checkEngagement(); err != nil {
		return "", err
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
