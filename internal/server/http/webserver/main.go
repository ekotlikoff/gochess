package main

import (
	"fmt"
	"gochess/internal/model"
	"strconv"
	"syscall/js"
)

var board js.Value
var pieceDragging js.Value
var originalTransform js.Value
var positionDragging, positionOriginal model.Position
var isMouseDown bool
var document = js.Global().Get("document")

func main() {
	done := make(chan struct{}, 0)
	// TODO init game in memory
	// TODO validate moves
	// TODO make http calls to interact with server
	// TODO only allow pieces of player color to be interacted with
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("subtract", js.FuncOf(subtract))
	elements := document.Call("getElementsByClassName", "piece")
	for i := 0; i < elements.Length(); i++ {
		elements.Index(i).Call(
			"addEventListener", "mousedown", js.FuncOf(mouseDown), false)
	}
	board = document.Call("getElementById", "board-layout-chessboard")
	document.Call("addEventListener", "mousemove", js.FuncOf(mouseMove), false)
	document.Call("addEventListener", "mouseup", js.FuncOf(mouseUp), false)
	<-done
}

func mouseDown(this js.Value, i []js.Value) interface{} {
	i[0].Call("preventDefault")
	isMouseDown = true
	pieceDragging = this
	rect := board.Call("getBoundingClientRect")
	width := rect.Get("right").Int() - rect.Get("left").Int()
	height := rect.Get("bottom").Int() - rect.Get("top").Int()
	squareWidth := width / 8
	squareHeight := height / 8
	x := i[0].Get("clientX").Int() - rect.Get("left").Int()
	y := i[0].Get("clientY").Int() - rect.Get("top").Int()
	gridX := x / squareWidth
	gridY := y / squareHeight
	positionDragging = model.Position{uint8(gridX), uint8(7 - gridY)}
	positionOriginal = model.Position{uint8(gridX), uint8(7 - gridY)}
	originalTransform = pieceDragging.Get("style").Get("transform")
	pieceDragging.Get("classList").Call("add", "dragging")
	return 0
}

func mouseMove(this js.Value, i []js.Value) interface{} {
	i[0].Call("preventDefault")
	if isMouseDown {
		rect := board.Call("getBoundingClientRect")
		width := rect.Get("right").Int() - rect.Get("left").Int()
		height := rect.Get("bottom").Int() - rect.Get("top").Int()
		squareWidth := width / 8
		squareHeight := height / 8
		x := i[0].Get("clientX").Int() - rect.Get("left").Int()
		if x > width {
			x = width
		} else if x < 0 {
			x = 0
		}
		y := i[0].Get("clientY").Int() - rect.Get("top").Int()
		if y > height {
			y = height
		} else if y < 0 {
			y = 0
		}
		gridX := x / squareWidth
		if gridX > 7 {
			gridX = 7
		} else if gridX < 0 {
			gridX = 0
		}
		gridY := y / squareHeight
		if gridY > 7 {
			gridY = 7
		} else if gridY < 0 {
			gridY = 0
		}
		positionDragging = model.Position{uint8(gridX), uint8(7 - gridY)}
		pieceWidth := pieceDragging.Get("clientWidth").Float()
		pieceHeight := pieceDragging.Get("clientHeight").Float()
		percentX := 100 * (float64(x) - pieceWidth/2) / float64(squareWidth)
		percentY := 100 * (float64(y) - pieceHeight/2) / float64(squareHeight)
		pieceDragging.Get("style").Set("transform",
			"translate("+fmt.Sprintf("%f%%,%f%%)", percentX, percentY))
	}
	return 0
}

func mouseUp(this js.Value, i []js.Value) interface{} {
	i[0].Call("preventDefault")
	if isMouseDown {
		newLocationClass := fmt.Sprintf("square-%d%d",
			positionDragging.File+1, positionDragging.Rank+1)
		elements := document.Call("getElementsByClassName", newLocationClass)
		elementsLength := elements.Length()
		for i := 0; i < elementsLength; i++ {
			elements.Index(i).Call("remove")
		}
		pieceDragging.Get("style").Set("transform", originalTransform)
		pieceDragging.Get("classList").Call("remove", "dragging")
		pieceDragging.Get("classList").Call("remove", fmt.Sprintf("square-%d%d",
			positionOriginal.File+1, positionOriginal.Rank+1))
		pieceDragging.Get("classList").Call("add", newLocationClass)
		pieceDragging = js.Undefined()
		isMouseDown = false
	}
	return 0
}

func add(this js.Value, i []js.Value) interface{} {
	value1 := document.Call("getElementById", i[0].String()).Get("value").String()
	value2 := document.Call("getElementById", i[1].String()).Get("value").String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)
	document.Call("getElementById", i[2].String()).Set("value", int1+int2)
	return 0
}

func subtract(this js.Value, i []js.Value) interface{} {
	value1 := document.Call("getElementById", i[0].String()).Get("value").String()
	value2 := document.Call("getElementById", i[1].String()).Get("value").String()
	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)
	document.Call("getElementById", i[2].String()).Set("value", int1-int2)
	return 0
}
