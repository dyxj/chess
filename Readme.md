# Chess

<!-- TOC -->
* [Chess](#chess)
  * [CLI](#cli)
  * [Road Map](#road-map)
<!-- TOC -->

## CLI
To run the game via CLI.
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

## Road Map
- [x] Engine
- [x] In person CLI
- [ ] Online Server(websockets) - base version
  - in progress: test cases
- [ ] Online Client - base version 
- [ ] Reconnect

