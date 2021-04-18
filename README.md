Build Requirements
* Install protoc https://grpc.io/docs/protoc-installation/
* Install golang
* Run make

TODO
* Detect all game end scenarios
    * https://www.chess.com/article/view/how-chess-games-can-end-8-ways-explained#:~:text=Agreement-,Win%2FLose%3A,%3A%20checkmate%2C%20resignation%20and%20timeout.
    - [x] Detect stalemate
    - [x] Handle pawn promotion
    - [ ] Insufficient material
    - [x] 50 move rule
    - [x] timeout
    - [x] resignation
    - [x] agreed draw
    - [x] Detect draw by repetition (same position 3 times)
        See https://en.wikipedia.org/wiki/Threefold_repetition for the definition
        of position.
        Let's do this with a map[positionId]uint8
        Just need a hash function to go from position -> positionId
        - [] Clear the state to save memory after an irreversible move (capture, pawn move, castle)
* Server
    - [x] http server
      - [x] Make a websocket server/client as an alternative to polling
      - [x] Move ttlmap out to a separate package that can be shared by http and websocket
      - [x] Move web client out of server directory to client directory as it will use both HTTP and websocket server depending on compile flags (or whatever)
        - [x] Restructure the webservers so that we have a front line webserver for auth and static files, and for other requests it proxies to a distinct websocket or http webserver
      - [x] Instead of hanging indefinitely on GET sync/async, return with no update after timeout
      - [x] Instead of hanging indefinitely on GET match, return after a server timeout with http 204
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
    - [x] Max matching time, after which we match the player with a chess engine (if connected)
* Client
    - [x] Golang WebAssembly web client
    - [] Ensure that webclient can enter matchmaking successfully after a gameover
    - [x] Show "pending draw" when requesting a draw
    - [x] Handle 202 responses from GET match while still pending
    - [x] Handle 200 response from GET sync/async with no update
    - [x] Display matched opponent name
    - [ ] Check cookies for session token instead of using hasSession bool
        - Not sure if possible, the golang cookiejar doesn't seem like it supports this.
    - [x] Display gameover results
    - [x] Request a draw/resign
    - [x] Display remaining time
    - [x] Display point advantage/captured pieces
    - [x] Play local match vs begin matchmaking
    - [x] Mobile support
    - [ ] Support pieceType options for pawn promotion
* General
    - [x] Travis CI test
    - [x] Fix race conditions
    - [x] Add race testing to travis CI
    - [x] Add vet to local build with https://github.com/grpc/grpc-go/blob/master/Makefile as inspiration
