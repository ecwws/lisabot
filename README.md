# lisabot

## Communication

* (A) Adapter
* (R) Responder
* (S) Lisa Bot Main Server

Adapter engage (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"command": {
		"id": "identifier",
		"action": "engage",
		"type": "adapter",
		"time": 1234567,
		"options": ["HMAC encoded source_identifier+time"]
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
		"type": "responder",
		"time": 1234567,
		"options": ["HMAC encoded source_identifier+time"]
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
	"message": {
		"message": "message",
		"from": "user_identifier",
		"room": "room_identifier"
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

