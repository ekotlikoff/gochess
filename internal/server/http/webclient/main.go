// +build webclient

package main

import (
	"github.com/Ekotlikoff/gochess/internal/model"
	"net/http"
	"net/http/cookiejar"
	"syscall/js"
)

func main() {
	done := make(chan struct{}, 0)
	game := model.NewGame()
	jar, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	clientModel := ClientModel{
		game: &game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"),
		matchingServerURI: "http://192.168.1.166:8000/",
		client:            client,
	}
	clientModel.initController(quiet)
	clientModel.initStyle()
	clientModel.viewInitBoard(clientModel.playerColor)
	<-done
}
