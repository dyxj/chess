# Chess

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
- [ ] Online(websockets)
  - in progress: changing websocket package to `ws` fixes bug of client receiving 1006 when read is interrupted 
- [ ] Online Client

