# Chess

<!-- TOC -->
* [Chess](#chess)
  * [Dev tools](#dev-tools)
    * [Taskfile](#taskfile)
    * [Postman](#postman)
  * [Road Map](#road-map)
  * [Modes](#modes)
    * [CLI: cmd.game-cli](#cli-cmdgame-cli)
<!-- TOC -->

## Dev tools

### Taskfile
It is recommended to install [Taskfile](https://taskfile.dev/) with autocompletion for ease of use.  
Alternatively you could look into the `_taskfiles` folder for commonly used commands.

```terminaloutput
chess git:(main) ✗ task
task: [default] task -l
task: Available tasks for this project:
* cli:build:       Build game-cli binary
* cli:run:         Run game-cli. Build binary if does not exist
```

### Postman
Postman collection provided in `postman` folder for convenience.

## Road Map
- [x] Engine
- [x] In person CLI
- [ ] Online Server(websockets) - base version
  - in progress: test cases
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

