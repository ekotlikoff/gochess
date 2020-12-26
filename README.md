TODO
* Detect all game end scenarios
    * https://www.chess.com/article/view/how-chess-games-can-end-8-ways-explained#:~:text=Agreement-,Win%2FLose%3A,%3A%20checkmate%2C%20resignation%20and%20timeout.
    - [x] Detect stalemate
    - [ ] Handle pawn promotion
    - [ ] Insufficient material
    - [ ] 50 move rule
    - [x] timeout
    - [x] resignation
    - [x] agreed draw
    - [ ] Detect draw by repetition (same position 3 times)
        See https://en.wikipedia.org/wiki/Threefold_repetition for the definition
        of position.
        Let's do this with a map[positionId]uint8
        Just need a hash function to go from position -> positionId
* Server
    - [x] http server
      - [ ] Instead of hanging on GET sync, return immediately with no update
      - [ ] Instead of hanging on GET match, return after 2 seconds so that we can have tighter timeouts
      - [ ] If no response from client in x seconds then call disconnect win for opponent
      - [ ] Support some mechanism for a user cancelling their matchmaking
      - [ ] GET /sync should also provide player and opponent's remaining time to keep client, server in sync
      - [ ] Implement /currentgame so that a disconnected client can reconnect
      - [ ] Use browser session storage to save the session token cookie, that way a client can refresh and check if their token is still valid/in a game https://developer.mozilla.org/en-US/docs/Web/API/Window/sessionStorage
      - [x] checkmate test
      - [x] requested draw test
      - [x] resignation test
      - [x] timeout test
    - [x] http server sessions
      - [x] testing
      - [ ] user auth?
    - [x] client agnostic matching server
        - [x] testing
* Client
    - [ ] Golang WebAssembly web client
        - [x] Show "pending draw" when requesting a draw
        - [ ] Handle 200 response from GET sync with no update
        - [x] Display matched opponent name
        - [ ] Check cookies for session token instead of using hasSession bool
            - Not sure if possible, the golang cookiejar doesn't seem like it supports this.
        - [x] Display gameover results
        - [x] Request a draw/resign
        - [x] Display remaining time
        - [x] Display point advantage/captured pieces
        - [x] Play local match vs begin matchmaking
        - [x] Mobile support
* General
    - [x] Travis CI test
