# Priscilla

[![Build Status](https://travis-ci.org/priscillachat/priscilla.svg?branch=master)](https://travis-ci.org/priscillachat/priscilla)
[![Code Climate](https://codeclimate.com/github/priscillachat/priscilla/badges/gpa.svg)](https://codeclimate.com/github/priscillachat/priscilla)

Priscilla is a chat bot written in go.

Most chat bots today are written in interpreted languages where
add-your-own-code is the typical model. The handful existing chatbot written in
go also took the same approach. In my opinion however, this isn't quite
practical for a compiled language like go, because any change to the bot would
require re-compiliation of the entire bot's codebase--something you can get away
with interpreted languages, but not so much in a compiled language. So I took a
different approach with Priscilla.

## Server-Client Model

Priscilla consists of four components (only first three are operational at the
moment):

* **Priscilla server**
* **Priscilla Adapter Client**
* **Priscilla Responder Client**
* **(Not Yet Implemented) Priscilla Persistent Storage Client**

### Priscilla Server

In reality, Priscilla server is just a message dispatcher server that
acts as the courier to facilite the communication between adapters (connect to
chat services) and responders (processes messages, performs actions, and
respond to messages). Communication protocol is in JSON over TCP and is
described later in the document. Though mostly stable, it's still in early
development stage and may change without notice until code base is stable.

Additionally, plan has been made to support a persistent storage client to
enable an unified persistent storage interface. The work on this is still
ongoing.

### Priscilla Adapter Client

Adapters are long running Priscilla clients that connects and listens on the chat
services then forward messages to Priscilla server, and listen for messages from
Priscilla server and forward them back to the chat service.

An unique feature about Priscilla, is that since Priscilla does not really
distinguish the connections trying to engage with her (yay, pun), and a new
unique source ID is assigned to an engaged adapter, there is no limit on how
many adapters can be active with a single Priscilla server at any time. You can
effectively have both HipChat, Slack, IRC, as well as a local shell test console
all connect to the same Priscilla server the same time, or connect multiple
HipChat/Slack/IRC organizations through multiple copies of the same adapters
to the same Priscilla Server.

Currently two adapters are functional:
* HipChat adapter: https://github.com/priscillachat/priscilla-hipchat
* Slack adapter https://github.com/priscillachat/priscilla-slack

Hipchat adapter supports full range of feature Priscilla supports. Slack adapter
at the moment only supports minimum set of functions to be functional. See
project page for more details.

Before Priscilla server would recognize the adapter, adapter has to first
"engage" the Priscilla server with engagement command. This would cause
Priscilla server to check and either confirm or assign a newly generated unique
source identifier for the message source. No messages would be forwarded if
engagement does not succeed.

### Priscilla Responder Client

Another unique feature about Priscilla is that because it communicates with
other components via JSON over TCP, there isn't any requirement that
Priscilla's plugin has to be written in go. Any program/process that can speak
Priscilla's communication protocol can act as an adapter or responder. In fact,
there's an entire subclass of responders called passive responder, does not
even require the responder to speak Priscilla's protocol at all. They are
literally commands Priscilla server would execute on the fly upon deteting a
matching pattern from the incoming message from an adapter and return the
output of the command back to the origin adapter. For example, the ping command
is simply implemented as a passive responder using unix "echo" command. (See
conf-example.yml)

There are two types of Responders **Active Responder** and **Passive Responder**

#### Active Responder

Like adapter, active responder is a long running process that listens for
requests form Priscilla server. And like adapter, active responder has to first
engage the Priscilla server. Then active responder would perform regex pattern
registration, where it tells Priscilla server what message should be forwarded to
it.

Active responder can also be used as an active trigger, where responses aren't
being triggered by incoming messages, instead, being triggered by timer event or
other event sources such as http or tcp events.

#### Passive Responder

This is a unique feature of Priscilla. A passive responder is essentially an
executable command that is specified in the Priscilla's config file what message
patter would trigger its execution and how it's executed. Once triggered, the
payload specified by config would be passed in as parameters, then the output of
the command would be captured and returned to the source adapter that triggered
the responder.

This is a powerful feature that makes both implementing and testing Priscilla
responders easy. You can literally write a single executable command that takes
input and produces output, and test it all without the need to have Priscilla
server running. When it's ready to be connected to Priscilla server, you
only have to specify in the config file how you want the command to be invoked.

For example, this is how you would implement a "ping" passive responder using
unix "echo" command, simply put in your Priscilla config file:

```yaml
responders:
  passive:
  - name: ping        # this is returned to adapter as source of the message
    match:
    - "^ping$"        # regular expression, you can specify multiple
    cmd: /bin/echo    # the command to be executed
    args: ["Pong"]    # the argument passed to the command
```

Passive responder can facilitate substring match substitution up to 10
submatches (0-9) enclosed in double-underscores. For example:

```yaml
responders:
  passive:
  - name: echo
    match:
    - "^echo (.+)$"
    cmd: /bin/echo
    args: ["__0__"]   # __0__ will be substituted by first submatch
```

Do be careful using the substitution, as it may have security concern. I would
recommend running Prescilla in a jailed environment (i.e. docker) to prevent
excape.

### Persistent Storage Client (Concept)

This would be used for facilitating any need for persistence storage in
Priscilla. The idea is to allow a new type of Priscilla client, Persistent
Storage Client, to register with Priscilla server and enable new type of command
query to be routed to it for storing and retrieving data from underlying storage
facilities.

## Configuration

The configuration file is in YAML format, and you would specify the
configuration file with **-conf** argument when starting Priscilla server.

```yaml
port: 4517    # default port for Priscilla server
prefix: pris  # default prefix
prefix-alt: [priscilla $mention cilla] # alternate prefix, not yet implemented
adapters:     # adapter could use these section for unified adapter config
  hipchat:
    params:
      user: "priscilla@priscilla.chat"
      pass: "abcdefg"
      nick: "Priscilla"
responders:
    passive:
    - name: echo
      match:
      - ^ping$
      noprefix: false # can be omitted, default is no activation without prefix
      cmd: /bin/echo
      args: ["pong"]
      fallthrough: false # can be omitted, if this is set to true, it will
                         # continue to match other patterns for activation,
                         # default behavior is to stop checking once it's
                         # activated once
    - name: cleverbot
      match:
      - ^(.*), pris$  # multiple activation patterns
      - ^(.*), pris\?$
      mentionmatch:   # match these patterns if it's being explictly mentioned
      - ^(.*)$
      noprefix: true  # will activate without prefix
      cmd: /usr/priscilla-scripts/cleverbot.sh # a script to curl cleverbot
      args: ["__0__"] # substitute with first submatch
    - name: wherami
      match:
      - ^whereami$
      cmd: /bin/echo
      args: ["I'm in __room__"] # priscilla substitute __room__ with room name
    - name: sha256
      match:
      - ^sha256 (.+)$
      cmd: /usr/priscilla-scripts/sha256.sh
      args: ["__0__"]
        # content of sha256.sh:
        # #!/bin/bash
        #
        # echo $1 | shasum -a 256 | cut -f 1 -d ' '
```

## Some background

Priscilla is a chat bot written purely in go. It all started when I started
learning go and wanted to do something useful and practical (um...chat bot is
useful and practical...right?) so I can practice the newly acquired go
knowledge. I tried to be as idiomatic go as this is a learning experience for me
to use go. I tried really hard to make the code free of locks and use the *share
memory by communicating* methodology. So far it's successful. Though that's not
to say I won't eventually stumble onto a problem that I couldn't solve using
this methodology. But at the moment, the entire Priscilla code base is
mutex-free and lock-free. If you're to write a new Priscilla adapter or
responder, I would recommend you to do the same, since mixing locks and channels
could increase the chance of getting deadlocks (reference: ??? I know I've read
about it somewhere, I just have to track down the article...)

At where I work, we make good use of our chat bot which is Lita bot
(https://www.lita.io/) based. I've written quite a few Lita plugins, some
are custom internal plugins strictly used within the company, some got published
as open source plugin. Before the Lita bot, we had a hubot, which I too
contributed to some internal plugins, but everybody in the team was tired of
writting coffeescript so it didn't take long before it was abondanded after the
lita bot came online.

Having worked with both Hubot and Lita bot, I've come to know them quite a bit
about them, including some of the internals. One thing common about both Hubot
and Lita bot as well several other bots out there is that they all work in a
"include" model, meaning when you want to run a bot, you essentially take a
copy of the bot code and add more custom code to it so they are "included" in
the new instance of the bot code. This model work pretty well with interpreted
languages because you always get a copy of the source code when you run anything
in the interpreted language. You're really not losing anything, nor adding any
overhead with the model.

One shortfall I can think of, though, is the fact it forces a very particular
way you can implement your bot's plugin. It forces not only communication
protocol, but also the implementation details.

With a compiled language, though, this model does not work well at all. As I
mentioned earlier that if the a bot written in go is to follow the same
"include" model, then you'll have to re-compile your code every time you make
modifications to your bot to add functionality. I think for majority of the
administrator of chat bots, that's something they'd shy away from. I would. To
get around the problem, I developed a model where routing is detached from the
message generator, so that all components can be individually seleted and
assembled and change in one wouldn't need to affect the change in another (as
long as the communication protocol stays the same).

## Communication

* (A) Adapter
* (R) Responder
* (S) Priscilla Main Server

### Adapter engage (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"action": "engage",
		"type": "adapter",
		"time": 123456789,
		"data": "base64(sha256-HMAC(unixtimestamp+source_identifier+secret))"
	}
}
```

### Active responder engage (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"action": "engage",
		"type": "responder"
		"time": 123456789,
		"data": "base64(sha256-HMAC(unixtimestamp+source_identifier+secret))"
	}
}
```

**Note**

To engage with Priscilla server, the client need to send in engagement
verification string in JSON format descirbed above. To calculate the string,
concatenate the unix time stamp (in seconds from epoc, as a string), source
identifier (the name you chose for the adapter), and the secret string that is
shared between Priscilla server and client. Run it through SHA256-HMAC, the
Base64 encoded result would be the data string to be sent over.

For example, a source with identifier "priscilla-slack", when engages at time
1474340021, with secret string being "abcdefghijkl", the pre-SHA256-HMAC encoded
string would be:

```
1474340021priscilla-slackabcdefghi
```

and the Base64 encoded SHA256-HMAC encoded string would be:

```
YtUmO0cxNkkelXwIHbotcCTrXb2R8sW+twBcelQ2NKA=
```

the engagement request would be:

```json
{
  "type": "command",
  "source": "priscilla-slack",
  "to": "server",
  "command": {
    "action": "engage",
    "type":" adapter",
    "time": 1474340021,
    "data": "YtUmO0cxNkkelXwIHbotcCTrXb2R8sW+twBcelQ2NKA="
  }
}
```

Timestamp must be within 10 seconds of server's current time. If the timestamp
is within range and HMAC matches the server's calculation, a "proceed" command
will be sent back as acknowledgement. If either time is out of range, or HMAC
calculation doesn't match, a "terminate" command will be sent back followed by
closing of the connection.

Other than the "type" field, adapter and responder engagement messages are
identical. However, Priscilla server treats adapter and responder differently.
Messages from Adapters, if a "to" field is ever set, it will be ignored as they
will always go through the pattern matching routine in the dispatcher. Messages
from responder on the other hand, would never go through a pattern matching
routine and would be sent directly to the "to" indicated source.

### Engagement success response (S->A, S->R)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "proceed",
		"data": "source_identifier",
	}
}
```

### Engagement failure response (S->A, S->R)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "terminate",
		"data": "error message",
	}
}
```

**Note**

When responder or adapter engage with Priscilla server, after the initial
engagement command is sent, client should wait for a command query. If it's an
command with action "proceed", then engagement is confirmed and normal
activities can begin. The source identifier that is assigned to the source will
be returned as the value in the "data" field. If the client's source id has
been accepted, it will be returned in this field. If it has been assigned a new
source identifier, it too will be returned in this field. Though this has very
little use for client as the server dispatcher would always ignore the source
field after initial engagement and insert the source id from the initial
engagement negation to every query came from that source.

If an error occures during the engagement process, then server would send back
a "terminate" command with an error message as the value in the "data" field,
then close the connection afterward.


### Disengage request (S->R/A, R/A->S)

```json
{
	"type": "command",
	"source": "identifier_or_server",
	"command": {
		"action": "disengage",
	}
}
```

### Active responder command registration (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "register",
		"type": "prefix",
		"data": "regex_string"
		"array": ["help-command", "help-message"],
		"options": ["fallthrough"]
	}
}
```

**Note**

"type" field: one of "prefix", "noprefix", "mention", "unhandled"

### Message from adapter (A->S)

**note:** "to" field can be left empty

**note:** "stripped" field is necessary because Adapters keeps it's own
reference of the mention name and it's not passed to the server. Server would
not know how to strip out the mention reference if it needs a stripped copy of
the message.

```json
{
	"type": "message",
	"source": "source_identifier",
	"to": "server",
	"message": {
		"message": "message",
		"from": "user_name",
		"room": "room_identifier",
		"mentioned": false,
		"stripped": "message stripped of mentions",
		"user": {
			"id": "id",
			"name": "user_name",
			"mention": "user_mention",
			"email": "user_email"
		}
	}
}
```

### Message from responder (R->S)

```json
{
	"type": "message",
	"source": "source_identifier",
	"to": "dest_identifier",
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier",
		"mentionnotify": ["user1", "user2", "user3"]
	}
}
```

### Request user information (R->A)

```json
{
	"type": "command",
	"source": "source_identifier",
	"to": "adapter_identifier",
	"command": {
		"id": "identifier",
		"action": "user_request",
		"type": "user / mention / email / id",
		"data": "username or mention or email"
	}
}
```

### User information response (A->R)

```json
{
	"type": "command",
	"source": "source_identifier",
	"to": "originator_identifier",
	"command": {
		"id": "identifier (use the identifier from the request)",
		"action": "info",
		"type": "user",
		"map": {"field1": "data1", "field2": "data2"}
	}
}
```

### Request room information (R->A)

```json
{
	"type": "command",
	"source": "source_identifier",
	"to": "adapter_identifier",
	"command": {
		"id": "identifier",
		"action": "room_request",
		"type": "name / id",
		"data": "room"
	}
}
```

### Room information response (A->R)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "info",
		"type": "room",
		"error": "Error message (if request fails)",
		"map": {"field1": "data1", "field2": "data2"}
	}
}
```

**Note** "error" field is only send back when error occurs. Information
requester should first evaluate whether "error" field is empty before
proceeding. If "error" field contains value, the map part of the information
response should be considered invalided and discarded.

**Note** Both user and room information request are neither validated nor
evaluated by Priscilla server, it's simply forwarded to the target adapter. The
information response from the adapter, is also directly forwarded to the
active responder, without being validated and evaluated. (Of course, it will
still go through the basic query validation before it's forwarded, if it fails
the basic validation, it will be discarded by the server). So the adapter and
the responder are actually responsible for validating the information request
and response.

**Note** "action": "info" is the only query from adapter that Priscilla server
would leave the "to" field intact. All other commands and messages from adapter
woudl always have "to" field emptied out.

## Fun stuff

The project name, Priscilla, which would be mostly referred as Pris in the
chatrooms, came from the famous novel Do Androids Dream of Electric Sheep? by
Philip K. Dick, some may know it as Blade Runner. Pris is one of the escaped
Nexus-6 androids that was hunted down by bounty hunters.
