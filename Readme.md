# Chess

## CLI version
```
cd cmd/game-cli
go build
./game-cli
```
In case chess symbols are not supported on your device, add `-icon number` to command.  
This uses numbers to represent pieces instead of symbols.

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

## To Do
- Add build scripts
- server to host multiple games
- UI, probably with ebitengine
- game timer

