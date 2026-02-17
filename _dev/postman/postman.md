# Chess Postman Collection

## HTTP Connection
`chess_http.postman_collection.json` includes HTTP calls with convenient pre and post scripts.
Pre and post scripts prefill code and tokens into postman environment variables.

## Websocket Connection
Unable to export websocket collection. After creating a websocket collection use the following url and payload.

### Connect Black
`ws://localhost:8080/room/connect?token={{btoken}}`
```json
{
    "type": "move",
    "payload": {
        "symbol": 1,
        "from": 48,
        "to": 40
    }
}
```
### Connect White
`ws://localhost:8080/room/connect?token={{wtoken}}`
```json
 {
  "type": "move",
  "payload": {
    "symbol": 1,
    "from": 11,
    "to": 19
  }
}
```