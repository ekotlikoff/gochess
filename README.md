TODO
* Detect all game end scenarios
    * https://www.chess.com/article/view/how-chess-games-can-end-8-ways-explained#:~:text=Agreement-,Win%2FLose%3A,%3A%20checkmate%2C%20resignation%20and%20timeout.
    - [x] Detect stalemate
    - [ ] Insufficient material
    - [ ] 50 move rule
    - [ ] timeout
    - [ ] resignation
    - [ ] agreed draw
    - [ ] Detect draw by repetition (same position 3 times)
        See https://en.wikipedia.org/wiki/Threefold_repetition for the definition
        of position.
        Let's do this with a map[positionId]uint8
        Just need a hash function to go from position -> positionId
* Server
    - [ ] http server
* Client
    - [ ] Golang WebAssembly web client
