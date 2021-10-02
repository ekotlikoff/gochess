### Todo
* Observability
    - [ ] More Prometheus metrics
    - [ ] Improve logging
* Server
    - [ ] Create http server objects to use for graceful shutdowns
      - use https://pkg.go.dev/net/http#Server.Shutdown to shut down without
        interrupting active connections
      - use https://pkg.go.dev/net/http#Server.RegisterOnShutdown (not sure if
        needed for websocket conns)
    - [ ] handle disconnected WS client gracefully, let matchserver know on disconnect and it starts a timer, ending game with timeout on alarm
    - [ ] Support some mechanism for a user cancelling their matchmaking
      - Probably would require a matching redesign, right now matchmaking is handled from a channel which doesn't support adhoc removal
    - http server sessions
      - [ ] user auth?
* Hosting
    - [ ] Restructure to enable other repos depending on this as a library
      - [ ] internal/server/frontend may need to go in pkg so that consuming
        repos can compile to lib.wasm? Or just commit the wasm binary?  What's
        the cleanest way of handling this?
      - [ ] Whatever we choose we should document clear in README.md for
        projects that consume this library
    - [ ] Port forwarding or public cloud?
    - [ ] Let's encrypt for SSL

### Done
* Observability
    - Prometheus metrics
        - [x] gauge for number of live games
        - [x] gauge for number of players matching
        - [x] counter for request count split by status
        - [x] histogram for latency split by status
    - [x] Centralize logging config into cmd/webserver/main
    - [x] Tracing
        - [x] Example trace
        - [x] Tracing for the websocket flow
* Detect all game end scenarios
    * https://www.chess.com/article/view/how-chess-games-can-end-8-ways-explained#:~:text=Agreement-,Win%2FLose%3A,%3A%20checkmate%2C%20resignation%20and%20timeout.
    - [x] Detect stalemate
    - [x] Handle pawn promotion
    - [x] 50 move rule
    - [x] Insufficient material
    - [x] timeout
    - [x] resignation
    - [x] agreed draw
    - [x] Detect draw by repetition (same position 3 times)
        See https://en.wikipedia.org/wiki/Threefold_repetition for the definition
        of position.
        Let's do this with a map[positionId]uint8
        Just need a hash function to go from position -> positionId
        - [x] Clear the state to save memory after an irreversible move (capture, pawn move, castle)
* Server
    - [x] http server
      - [x] Websocket ping + pong messages at certain interval (if no other message written), this lets us have a read and write deadline enabling us to close out the read loops cleanly
      - [x] Make a websocket server/client as an alternative to polling
      - [x] Move ttlmap out to a separate package that can be shared by http and websocket
      - [x] Move web client out of server directory to client directory as it will use both HTTP and websocket server depending on compile flags (or whatever)
        - [x] Restructure the webservers so that we have a front line webserver for auth and static files, and for other requests it proxies to a distinct websocket or http webserver
      - [x] Instead of hanging indefinitely on GET sync/async, return with no update after timeout
      - [x] Instead of hanging indefinitely on GET match, return after a server timeout with http 204
      - [x] checkmate test
      - [x] requested draw test
      - [x] resignation test
      - [x] timeout test
      - [x] GET /sync should also provide player and opponent's remaining time to keep client, server in sync (implemented for WS only)
      - [x] Implement GET /session so that a disconnected client can reconnect
        - On initial page load client calls /session
        - Server will return the user's username if supplied a valid token
        - Server will also check for an ongoing game, if one is ongoing it sends the game state to the client, client will set the game state accordingly and rejoin the game
        - [x] Redesign websocket to gracefully handle an initial connection + match request and a reconnect
        - [x] test websocket reconnect, and add a lock for each session.  We should only process one request from a session at a time (WS or http)
        - [x] Implement reconnect logic in web client
        - [x] Redesign clientDoneWithMatch, instead the match signifies when it's over and client servers synchronize on that (WaitForMatchOver), then resetting themselves.  The other way around required client servers to let the matchserver know when they were done with a match, this made less sense because a client could have disconnected and take some time to reconnect, it's harder to tell when they're done with the match.  Worth noting here that to scale we would likely want to be somewhat aggressive on reeping inactive players in sessionCache, because players could frequently disconnect when losing without formally resigning and letting their client server reset the player object, thus keeping a reference to the match and preventing GC.
    - [x] http server sessions
      - [x] testing
    - [x] client agnostic matching server
        - [x] testing
        - [x] if client abandoned set the opponent as the winner
          - [x] websocket
          - [x] engine
    - [x] Max matching time, after which we match the player with a chess engine (if connected)
* Client
    - [x] Golang WebAssembly web client
    - [x] Ensure that webclient can enter matchmaking successfully after a gameover
    - [x] Show "pending draw" when requesting a draw
    - [x] Handle 202 responses from GET match while still pending
    - [x] Handle 200 response from GET sync/async with no update
    - [x] Display matched opponent name
    - [x] Display gameover results
    - [x] Request a draw/resign
    - [x] Display remaining time
    - [x] Display point advantage/captured pieces
    - [x] Play local match vs begin matchmaking
    - [x] Mobile support
* General
    - [x] Travis CI test
    - [x] Fix race conditions
    - [x] Add race testing to travis CI
    - [x] Add vet to local build with https://github.com/grpc/grpc-go/blob/master/Makefile as inspiration
    - [x] Support pieceType options for pawn promotion
