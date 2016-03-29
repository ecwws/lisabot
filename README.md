# lisabot

## Communication

Adapter engagement (A->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"id": "identifier",
	"command": {
		"action": "engage",
		"type": "adapter",
		"time": 1234567,
		"options": ["HMAC encoded source_identifier+time"]
	}
}
```

Active interceptor engagement (I->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"id": "identifier",
	"command": {
		"action": "engage",
		"type": "interceptor",
		"time": 1234567,
		"options": ["HMAC encoded source_identifier+time"]
	}
}
```

Active interceptor command registration (I->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"id": "identifier",
	"command": {
		"action": "register",
		"type": "command",
		"pattern": "regex_string"
		"options": ["fallthrough", "unhandled"]
	}
}
```

Active interceptor command registration complete notification (I->S)

```json
{
	"type": "command",
	"source": "source_identifier",
	"id": "identifier",
	"command": {
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
		"direct": false,
		"room": "room_identifier"
	}
}
```

Request user information (S->A)

```json
{
	"type": "command",
	"source": "server",
	"id": "identifier",
	"command": {
		"action": "request",
		"type": "user",
		"options": ["user1", "user2", "user3", "user4"]
	}
}
```

Request room information (S->A)

```json
{
	"type": "command",
	"source": "server",
	"id": "identifier",
	"command": {
		"action": "request",
		"type": "room",
		"options": ["room1", "room2", "room3", "room4"]
	}
}
```

Disengage request (S->I/A, I/A->S)

```json
{
	"type": "command",
	"source": "identifier_or_server",
	"id": "identitfier",
	"command": {
		"action": "disengage",
	}
}
```

