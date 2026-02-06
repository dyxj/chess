# Chess

## CLI version
```
cd cmd/game-cli
go build
./game-cli
```
In case chess symbols are not supported on for device add `-icon number` to command.  
This uses numbers to represent pieces instead of symbols.

## To Do
- move converter with rank and file, provides easier interface for simpler agents
- Determining check mate, would not be great to have to generate all moves on check
- Is there a need to keep track of check?
- Add build scripts

