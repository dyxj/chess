# Chess

<!-- TOC -->
* [Chess](#chess)
  * [Dev tools](#dev-tools)
    * [Taskfile](#taskfile)
    * [Postman](#postman)
  * [Road Map](#road-map)
  * [Modes](#modes)
    * [CLI: cmd.game-cli](#cli-cmdgame-cli)
  * [Server: cmd.game-server](#server-cmdgame-server)
    * [APIs](#apis)
      * [Create Room](#create-room)
      * [Join Room](#join-room)
      * [Connect](#connect)
        * [Action](#action)
        * [Events](#events)
<!-- TOC -->

## Dev tools

### Taskfile
It is recommended to install [Taskfile](https://taskfile.dev/) with autocompletion for ease of use.  
Alternatively you could look into the `_taskfiles` folder for commonly used commands.

```terminaloutput
➜  chess git:(main) ✗ task
task: [default] task -l
task: Available tasks for this project:
* cli:build:              Build game-cli binary
* cli:run:                Run game-cli. Build binary if does not exist
* server:build:           Build game-server binary
* server:run-local:       Run game-server. Build binary if does not exist
* test:all:               Run all tests
* test:coverage:          Run all tests with cross package coverage
```

### Postman
Postman collection provided in `_dev/postman` folder for convenience.

## Road Map
- [x] Engine
- [x] In person CLI
- [x] Online Server(websockets) - base version
- [ ] Online Client - base version
- [ ] Reconnect

## Modes

### CLI: cmd.game-cli
`cli` is a simple command line interface for playing chess in person.  
Run To run the game via CLI.
```
task cli:run
```
In case chess symbols are not supported on your device.
Run
```shell
task cli:run icon=number
```

```shell
---------------------------
8 | ♜| ♞| ♝| ♛| ♚| ♝| ♞| ♜|
---------------------------
7 | ♟| ♟| ♟| ♟| ♟| ♟| ♟| ♟|
---------------------------
6 | ·| ·| ·| ·| ·| ·| ·| ·|
---------------------------
5 | ·| ·| ·| ·| ·| ·| ·| ·|
---------------------------
4 | ·| ·| ·| ·| ·| ·| ·| ·|
---------------------------
3 | ·| ·| ·| ·| ·| ·| ·| ·|
---------------------------
2 |-♙|-♙|-♙|-♙|-♙|-♙|-♙|-♙|
---------------------------
1 |-♖|-♘|-♗|-♕|-♔|-♗|-♘|-♖|
---------------------------
    a  b  c  d  e  f  g  h
```

## Server: cmd.game-server
`server` chess server for online play.
Uses websockets for communication.

### APIs

#### Create Room
```http request
POST /room
No body required

Response:
{
    "code": "MTOTQF",
    "status": "waiting",
    "createdTime": "2026-02-23T12:47:45.780779+01:00"
}
```

#### Join Room
```http request
POST /room/{code}/join
Body:
{
  "name": "bplayer",
  "color": "black"
}

Response:
{
    "token": "884f19663e9342fddbb3ee9499a11f90"
}
```

#### Connect
```http request
ws://localhost:8080/room/connect?token={token}
```
After connection is established actions can be sent via websocket in the following format:  
##### Action
**Move**
```json
{
  "type": "move",
  "payload": {
    "symbol": 1,
    "from": 48,
    "to": 50,
    "promotion": 0
  }
}
```

##### Events
Events are sent by the server in the following format:  
**Message**
```json
{
  "type": "message",
  "payload": {
    "message": "Waiting for white player"
  }
}
```

**Round**  
Describes current round state.
```json
{
  "type": "round",
  "payload": {
    "count": 0,
    "state": "in_progress",
    "grid": [
      4,  2,  3,  5,  6,  3,  2,  4,
      1,  1,  1,  1,  1,  1,  1,  1,
      0,  0,  0,  0,  0,  0,  0,  0,
      0,  0,  0,  0,  0,  0,  0,  0,
      0,  0,  0,  0,  0,  0,  0,  0,
      0,  0,  0,  0,  0,  0,  0,  0,
     -1, -1, -1, -1, -1, -1, -1, -1,
     -4, -2, -3, -5, -6, -3, -2, -4
    ],
    "activeColor": "white"
  }
}}
```

**Error**
```json
{
  "type": "error",
  "payload": {
    "error": "invalid symbol"
  }
}
```

**Resign**
```json
{
  "type": "resign",
  "payload": {
    "resigner": "black",
    "winner": "white"
  }
}
```