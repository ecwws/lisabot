# Priscilla

Priscilla is a chat bot written in go. Unlike other popular chat bots, which are
written in interpreted language, in my opinion, it's a bit unpractical to write
a chat bot in go that requires source code modification or re-compilation every
time you want to add functions to the bot. So I took different approach.

Priscilla consistes three different components: *Priscilla server*, *Priscilla
Adapter*, and *Priscilla Responder*.

In reality, Priscilla server is actually just a message dispatcher that
acts as the middle man for adapters (connect to chat services) and responders
(processes messages, performs actions, and respond to messages) to communicate
with each other. Communication protocol is in JSON over TCP and is described
later in the document.

One unique feature about Priscilla is that because it communicates with other
components via JSON over TCP, there isn't any requirement that Priscilla's
plugin has to be written in go. Any program/process that can speak Priscilla's
communication protocol can act as an adapter or responder. In fact, there's an
entire subclass of responders called passive responder, does not even require
the responder to speak Priscilla's protocol at all. They are literally commands
Priscilla server would execute on the fly upon deteting a matching pattern from
the incoming message from an adapter and return the output back to the origin
adapter. That's why there isn't an internal "ping" command, instead, you would
just simply implement it as a passive responder using unix "echo" command.

Another unique feature about Priscilla, is that since Priscilla does not really
distinguish the connections trying to engage with her (yay, pun), there is no
limit on how many adapters can be active with a single Priscilla server at the
same time. You can have both HipChat and Slack connect to the same Priscilla
server the same time, or connect multiple HipChat/Slack organizations through
multiple adapters that all connects to the same Priscilla Server.

## About

Priscilla is a chat bot written purely in go. It all started when I started
learning go and wanted to do something useful and practical (um...chat bot is
useful and practical...right?) so I can practice the newly acquired go
knowledge.

At where I work, we make good use of our chat bot, a Lita bot
https://www.lita.io/. I've written quite a few Lita plugins, some
are custom internal plugins strictly used within the company, some got published
as open source plugin. Before the Lita bot, we had a hubot, which I too
contributed to some internal plugins.

Having worked with both Hubot and Lita bot, I've come to know them quite well
including some of the internals. Both Hubot and Lita bot as well as many other
popular bots out there works in a "include" model, where you want to run a bot,
you would take a copy of the bot code and add more custom code to it so they
are "included" in the new instance of the bot. The shortfall of this model, is
the fact it forces a very particular way you can implement your bot's plugin. It
works well most time but once a while certain problem arises that I just with I
had a bit more freedom in implementing the plugin. Of course, on top of that,
this model would work well only for bots are written in interpreted
language--most bot users/admins probably don't want to re-compile the bot every
time they add a feature/plugin. Thus  the Server/Client model is developed.

## Adapters

Adapters are long running Priscilla clients that connects and listens on the chat
services then forward messages to Priscilla server, and listen for messages from
Priscilla server and forward them to the chat service.

Currently only one adapter is functional, the HipChat adapter:
https://github.com/priscillachat/priscilla-hipchat

Before Priscilla can recognize the adapter, adapter has to first "engage" the
Priscilla server. No messages would be forwarded if engagement did not succeed.

## Responders

Responders are plugins that would respond to messages from adapters, and/or
perform specific defined actions. There are two type of responders: active
responder and passive responder.

### Active Responder

Like adapter, active responder is a long running process that listens for
requests form Priscilla server. And like adapter, active responder has to first
engage the Priscilla server. Then active responder would perform pattern
registration, where it tells Priscilla server what message should be forwarded to
it.

### Passive Responder

This is a unique feature of Priscilla. A passive responder is essentially an
executable command that is specified in the Priscilla's config file what message
patter would trigger its execution and how it's executed. Then the output of
the command would be returned to the source adapter that triggered the
responder.

This is a powerful feature that makes both implementing and testing Priscilla
responders easy. You can literally write a single executable command that takes
input and produces output, and test it all without the need to have Priscilla
being present at all. When time comes to connect the responder to Priscilla, you
only have to specify in the config file how you want the command invoked.

For example, this is how you would implement a "ping" passive responder using
unix "echo" command, simply put in your Priscilla config file:

```yaml
responders:
  passive:
  - match: ^ping$
    cmd: /bin/echo
    args: ["Pong"]
```

There, you just implemented a "ping" responder. There are so many things you can
do by writing config file to utilize existing programs without writing any code
at all! Imagine that.

## Communication

* (A) Adapter
* (R) Responder
* (S) Priscilla Main Server

Adapter engage (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "engage",
		"type": "adapter"
	}
}
```

Active responder engage (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "engage",
		"type": "responder"
	}
}
```

Active responder command registration (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "register",
		"type": "pattern",
		"data": "regex_string"
		"options": ["fallthrough", "unhandled"]
	}
}
```

Active responder command registration complete notification (R->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "noop",
		"type": "command",
	}
}
```

Message from adapter (A->S)

```json
{
	"type": "message",
	"source": "source_identifier",
  "to": "server",
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier",
    "mentioned": boolean
	}
}
```

Message from active responder (R->S)

```json
{
	"type": "message",
	"source": "source_identifier",
  "to": "dest_identifier",
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier",
    "mentioned": boolean
	}
}
```

Request user information (S->A)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "request",
		"type": "user",
		"options": ["user1", "user2", "user3", "user4"]
	}
}
```

User information response (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "info",
		"type": "user",
		"array": ["data1", "data2", "data3", "data4"]
	}
}
```

Request room information (S->A)

```json
{
	"type": "command",
	"source": "server",
	"command": {
		"id": "identifier",
		"action": "request",
		"type": "room",
		"options": ["room1", "room2", "room3", "room4"]
	}
}
```

Room information response (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "info",
		"type": "room",
		"array": ["data1", "data2", "data3", "data4"]
	}
}
```

Disengage request (S->R/A, R/A->S)

```json
{
	"type": "command",
	"source": "identifier_or_server",
	"command": {
		"id": "identitfier",
		"action": "disengage",
	}
}
```

