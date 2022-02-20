// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: api/chessengine.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PromotePiece_Piece int32

const (
	PromotePiece_NONE   PromotePiece_Piece = 0
	PromotePiece_QUEEN  PromotePiece_Piece = 1
	PromotePiece_ROOK   PromotePiece_Piece = 2
	PromotePiece_BISHOP PromotePiece_Piece = 3
	PromotePiece_KNIGHT PromotePiece_Piece = 4
)

// Enum value maps for PromotePiece_Piece.
var (
	PromotePiece_Piece_name = map[int32]string{
		0: "NONE",
		1: "QUEEN",
		2: "ROOK",
		3: "BISHOP",
		4: "KNIGHT",
	}
	PromotePiece_Piece_value = map[string]int32{
		"NONE":   0,
		"QUEEN":  1,
		"ROOK":   2,
		"BISHOP": 3,
		"KNIGHT": 4,
	}
)

func (x PromotePiece_Piece) Enum() *PromotePiece_Piece {
	p := new(PromotePiece_Piece)
	*p = x
	return p
}

func (x PromotePiece_Piece) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PromotePiece_Piece) Descriptor() protoreflect.EnumDescriptor {
	return file_api_chessengine_proto_enumTypes[0].Descriptor()
}

func (PromotePiece_Piece) Type() protoreflect.EnumType {
	return &file_api_chessengine_proto_enumTypes[0]
}

func (x PromotePiece_Piece) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PromotePiece_Piece.Descriptor instead.
func (PromotePiece_Piece) EnumDescriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{1, 0}
}

type AsyncRequest_Type int32

const (
	AsyncRequest_DRAW   AsyncRequest_Type = 0
	AsyncRequest_RESIGN AsyncRequest_Type = 1
)

// Enum value maps for AsyncRequest_Type.
var (
	AsyncRequest_Type_name = map[int32]string{
		0: "DRAW",
		1: "RESIGN",
	}
	AsyncRequest_Type_value = map[string]int32{
		"DRAW":   0,
		"RESIGN": 1,
	}
)

func (x AsyncRequest_Type) Enum() *AsyncRequest_Type {
	p := new(AsyncRequest_Type)
	*p = x
	return p
}

func (x AsyncRequest_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AsyncRequest_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_api_chessengine_proto_enumTypes[1].Descriptor()
}

func (AsyncRequest_Type) Type() protoreflect.EnumType {
	return &file_api_chessengine_proto_enumTypes[1]
}

func (x AsyncRequest_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AsyncRequest_Type.Descriptor instead.
func (AsyncRequest_Type) EnumDescriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{4, 0}
}

type GameStart_Color int32

const (
	GameStart_BLACK GameStart_Color = 0
	GameStart_WHITE GameStart_Color = 1
)

// Enum value maps for GameStart_Color.
var (
	GameStart_Color_name = map[int32]string{
		0: "BLACK",
		1: "WHITE",
	}
	GameStart_Color_value = map[string]int32{
		"BLACK": 0,
		"WHITE": 1,
	}
)

func (x GameStart_Color) Enum() *GameStart_Color {
	p := new(GameStart_Color)
	*p = x
	return p
}

func (x GameStart_Color) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GameStart_Color) Descriptor() protoreflect.EnumDescriptor {
	return file_api_chessengine_proto_enumTypes[2].Descriptor()
}

func (GameStart_Color) Type() protoreflect.EnumType {
	return &file_api_chessengine_proto_enumTypes[2]
}

func (x GameStart_Color) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GameStart_Color.Descriptor instead.
func (GameStart_Color) EnumDescriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{6, 0}
}

type GameMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Request:
	//	*GameMessage_ChessMove
	//	*GameMessage_AsyncRequest
	//	*GameMessage_GameStart
	Request isGameMessage_Request `protobuf_oneof:"Request"`
}

func (x *GameMessage) Reset() {
	*x = GameMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameMessage) ProtoMessage() {}

func (x *GameMessage) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameMessage.ProtoReflect.Descriptor instead.
func (*GameMessage) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{0}
}

func (m *GameMessage) GetRequest() isGameMessage_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (x *GameMessage) GetChessMove() *ChessMove {
	if x, ok := x.GetRequest().(*GameMessage_ChessMove); ok {
		return x.ChessMove
	}
	return nil
}

func (x *GameMessage) GetAsyncRequest() *AsyncRequest {
	if x, ok := x.GetRequest().(*GameMessage_AsyncRequest); ok {
		return x.AsyncRequest
	}
	return nil
}

func (x *GameMessage) GetGameStart() *GameStart {
	if x, ok := x.GetRequest().(*GameMessage_GameStart); ok {
		return x.GameStart
	}
	return nil
}

type isGameMessage_Request interface {
	isGameMessage_Request()
}

type GameMessage_ChessMove struct {
	ChessMove *ChessMove `protobuf:"bytes,1,opt,name=chess_move,json=chessMove,proto3,oneof"`
}

type GameMessage_AsyncRequest struct {
	AsyncRequest *AsyncRequest `protobuf:"bytes,2,opt,name=async_request,json=asyncRequest,proto3,oneof"`
}

type GameMessage_GameStart struct {
	GameStart *GameStart `protobuf:"bytes,3,opt,name=game_start,json=gameStart,proto3,oneof"`
}

func (*GameMessage_ChessMove) isGameMessage_Request() {}

func (*GameMessage_AsyncRequest) isGameMessage_Request() {}

func (*GameMessage_GameStart) isGameMessage_Request() {}

type PromotePiece struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Piece PromotePiece_Piece `protobuf:"varint,1,opt,name=piece,proto3,enum=rustchess.PromotePiece_Piece" json:"piece,omitempty"`
}

func (x *PromotePiece) Reset() {
	*x = PromotePiece{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PromotePiece) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PromotePiece) ProtoMessage() {}

func (x *PromotePiece) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PromotePiece.ProtoReflect.Descriptor instead.
func (*PromotePiece) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{1}
}

func (x *PromotePiece) GetPiece() PromotePiece_Piece {
	if x != nil {
		return x.Piece
	}
	return PromotePiece_NONE
}

type Position struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	File uint32 `protobuf:"varint,1,opt,name=file,proto3" json:"file,omitempty"`
	Rank uint32 `protobuf:"varint,2,opt,name=rank,proto3" json:"rank,omitempty"`
}

func (x *Position) Reset() {
	*x = Position{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Position) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Position) ProtoMessage() {}

func (x *Position) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Position.ProtoReflect.Descriptor instead.
func (*Position) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{2}
}

func (x *Position) GetFile() uint32 {
	if x != nil {
		return x.File
	}
	return 0
}

func (x *Position) GetRank() uint32 {
	if x != nil {
		return x.Rank
	}
	return 0
}

type ChessMove struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OriginalPosition *Position     `protobuf:"bytes,1,opt,name=original_position,json=originalPosition,proto3" json:"original_position,omitempty"`
	NewPosition      *Position     `protobuf:"bytes,2,opt,name=new_position,json=newPosition,proto3" json:"new_position,omitempty"`
	PromotePiece     *PromotePiece `protobuf:"bytes,3,opt,name=promote_piece,json=promotePiece,proto3" json:"promote_piece,omitempty"`
}

func (x *ChessMove) Reset() {
	*x = ChessMove{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChessMove) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChessMove) ProtoMessage() {}

func (x *ChessMove) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChessMove.ProtoReflect.Descriptor instead.
func (*ChessMove) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{3}
}

func (x *ChessMove) GetOriginalPosition() *Position {
	if x != nil {
		return x.OriginalPosition
	}
	return nil
}

func (x *ChessMove) GetNewPosition() *Position {
	if x != nil {
		return x.NewPosition
	}
	return nil
}

func (x *ChessMove) GetPromotePiece() *PromotePiece {
	if x != nil {
		return x.PromotePiece
	}
	return nil
}

type AsyncRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type AsyncRequest_Type `protobuf:"varint,1,opt,name=type,proto3,enum=rustchess.AsyncRequest_Type" json:"type,omitempty"`
}

func (x *AsyncRequest) Reset() {
	*x = AsyncRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AsyncRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AsyncRequest) ProtoMessage() {}

func (x *AsyncRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AsyncRequest.ProtoReflect.Descriptor instead.
func (*AsyncRequest) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{4}
}

func (x *AsyncRequest) GetType() AsyncRequest_Type {
	if x != nil {
		return x.Type
	}
	return AsyncRequest_DRAW
}

type GameTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlayerMainTime uint32 `protobuf:"varint,1,opt,name=player_main_time,json=playerMainTime,proto3" json:"player_main_time,omitempty"`
	Increment      uint32 `protobuf:"varint,2,opt,name=increment,proto3" json:"increment,omitempty"`
}

func (x *GameTime) Reset() {
	*x = GameTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameTime) ProtoMessage() {}

func (x *GameTime) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameTime.ProtoReflect.Descriptor instead.
func (*GameTime) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{5}
}

func (x *GameTime) GetPlayerMainTime() uint32 {
	if x != nil {
		return x.PlayerMainTime
	}
	return 0
}

func (x *GameTime) GetIncrement() uint32 {
	if x != nil {
		return x.Increment
	}
	return 0
}

type GameStart struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PlayerColor    GameStart_Color `protobuf:"varint,1,opt,name=player_color,json=playerColor,proto3,enum=rustchess.GameStart_Color" json:"player_color,omitempty"`
	PlayerGameTime *GameTime       `protobuf:"bytes,2,opt,name=player_game_time,json=playerGameTime,proto3" json:"player_game_time,omitempty"`
}

func (x *GameStart) Reset() {
	*x = GameStart{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_chessengine_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameStart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameStart) ProtoMessage() {}

func (x *GameStart) ProtoReflect() protoreflect.Message {
	mi := &file_api_chessengine_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameStart.ProtoReflect.Descriptor instead.
func (*GameStart) Descriptor() ([]byte, []int) {
	return file_api_chessengine_proto_rawDescGZIP(), []int{6}
}

func (x *GameStart) GetPlayerColor() GameStart_Color {
	if x != nil {
		return x.PlayerColor
	}
	return GameStart_BLACK
}

func (x *GameStart) GetPlayerGameTime() *GameTime {
	if x != nil {
		return x.PlayerGameTime
	}
	return nil
}

var File_api_chessengine_proto protoreflect.FileDescriptor

var file_api_chessengine_proto_rawDesc = []byte{
	0x0a, 0x15, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x68, 0x65, 0x73, 0x73, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65,
	0x73, 0x73, 0x22, 0xc6, 0x01, 0x0a, 0x0b, 0x47, 0x61, 0x6d, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x35, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x73, 0x73, 0x5f, 0x6d, 0x6f, 0x76, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65,
	0x73, 0x73, 0x2e, 0x43, 0x68, 0x65, 0x73, 0x73, 0x4d, 0x6f, 0x76, 0x65, 0x48, 0x00, 0x52, 0x09,
	0x63, 0x68, 0x65, 0x73, 0x73, 0x4d, 0x6f, 0x76, 0x65, 0x12, 0x3e, 0x0a, 0x0d, 0x61, 0x73, 0x79,
	0x6e, 0x63, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x41, 0x73, 0x79,
	0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x61, 0x73, 0x79,
	0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x35, 0x0a, 0x0a, 0x67, 0x61, 0x6d,
	0x65, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e,
	0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x48, 0x00, 0x52, 0x09, 0x67, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x42, 0x09, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x83, 0x01, 0x0a, 0x0c,
	0x50, 0x72, 0x6f, 0x6d, 0x6f, 0x74, 0x65, 0x50, 0x69, 0x65, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x05,
	0x70, 0x69, 0x65, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x72, 0x75,
	0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x6f, 0x74, 0x65, 0x50,
	0x69, 0x65, 0x63, 0x65, 0x2e, 0x50, 0x69, 0x65, 0x63, 0x65, 0x52, 0x05, 0x70, 0x69, 0x65, 0x63,
	0x65, 0x22, 0x3e, 0x0a, 0x05, 0x50, 0x69, 0x65, 0x63, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x4f,
	0x4e, 0x45, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x51, 0x55, 0x45, 0x45, 0x4e, 0x10, 0x01, 0x12,
	0x08, 0x0a, 0x04, 0x52, 0x4f, 0x4f, 0x4b, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x42, 0x49, 0x53,
	0x48, 0x4f, 0x50, 0x10, 0x03, 0x12, 0x0a, 0x0a, 0x06, 0x4b, 0x4e, 0x49, 0x47, 0x48, 0x54, 0x10,
	0x04, 0x22, 0x32, 0x0a, 0x08, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x66, 0x69, 0x6c,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x61, 0x6e, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x04, 0x72, 0x61, 0x6e, 0x6b, 0x22, 0xc3, 0x01, 0x0a, 0x09, 0x43, 0x68, 0x65, 0x73, 0x73, 0x4d,
	0x6f, 0x76, 0x65, 0x12, 0x40, 0x0a, 0x11, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x5f,
	0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x50, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x10, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x50, 0x6f, 0x73,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x0c, 0x6e, 0x65, 0x77, 0x5f, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x72, 0x75,
	0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x0b, 0x6e, 0x65, 0x77, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3c, 0x0a,
	0x0d, 0x70, 0x72, 0x6f, 0x6d, 0x6f, 0x74, 0x65, 0x5f, 0x70, 0x69, 0x65, 0x63, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x6f, 0x74, 0x65, 0x50, 0x69, 0x65, 0x63, 0x65, 0x52, 0x0c, 0x70,
	0x72, 0x6f, 0x6d, 0x6f, 0x74, 0x65, 0x50, 0x69, 0x65, 0x63, 0x65, 0x22, 0x5e, 0x0a, 0x0c, 0x41,
	0x73, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x72, 0x75, 0x73, 0x74,
	0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x41, 0x73, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x1c, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x52, 0x41, 0x57, 0x10, 0x00, 0x12,
	0x0a, 0x0a, 0x06, 0x52, 0x45, 0x53, 0x49, 0x47, 0x4e, 0x10, 0x01, 0x22, 0x52, 0x0a, 0x08, 0x47,
	0x61, 0x6d, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x70, 0x6c, 0x61, 0x79, 0x65,
	0x72, 0x5f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0e, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x4d, 0x61, 0x69, 0x6e, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x63, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x69, 0x6e, 0x63, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x22,
	0xa8, 0x01, 0x0a, 0x09, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x3d, 0x0a,
	0x0c, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e,
	0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x72, 0x74, 0x2e, 0x43, 0x6f, 0x6c, 0x6f, 0x72, 0x52,
	0x0b, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6c, 0x6f, 0x72, 0x12, 0x3d, 0x0a, 0x10,
	0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x67, 0x61, 0x6d, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65,
	0x73, 0x73, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x0e, 0x70, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x47, 0x61, 0x6d, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x1d, 0x0a, 0x05, 0x43,
	0x6f, 0x6c, 0x6f, 0x72, 0x12, 0x09, 0x0a, 0x05, 0x42, 0x4c, 0x41, 0x43, 0x4b, 0x10, 0x00, 0x12,
	0x09, 0x0a, 0x05, 0x57, 0x48, 0x49, 0x54, 0x45, 0x10, 0x01, 0x32, 0x49, 0x0a, 0x09, 0x52, 0x75,
	0x73, 0x74, 0x43, 0x68, 0x65, 0x73, 0x73, 0x12, 0x3c, 0x0a, 0x04, 0x47, 0x61, 0x6d, 0x65, 0x12,
	0x16, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2e, 0x47, 0x61, 0x6d, 0x65,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x72, 0x75, 0x73, 0x74, 0x63, 0x68,
	0x65, 0x73, 0x73, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22,
	0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x45, 0x6b, 0x6f, 0x74, 0x6c, 0x69, 0x6b, 0x6f, 0x66, 0x66, 0x2f, 0x67,
	0x6f, 0x63, 0x68, 0x65, 0x73, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_chessengine_proto_rawDescOnce sync.Once
	file_api_chessengine_proto_rawDescData = file_api_chessengine_proto_rawDesc
)

func file_api_chessengine_proto_rawDescGZIP() []byte {
	file_api_chessengine_proto_rawDescOnce.Do(func() {
		file_api_chessengine_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_chessengine_proto_rawDescData)
	})
	return file_api_chessengine_proto_rawDescData
}

var file_api_chessengine_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_api_chessengine_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_api_chessengine_proto_goTypes = []interface{}{
	(PromotePiece_Piece)(0), // 0: rustchess.PromotePiece.Piece
	(AsyncRequest_Type)(0),  // 1: rustchess.AsyncRequest.Type
	(GameStart_Color)(0),    // 2: rustchess.GameStart.Color
	(*GameMessage)(nil),     // 3: rustchess.GameMessage
	(*PromotePiece)(nil),    // 4: rustchess.PromotePiece
	(*Position)(nil),        // 5: rustchess.Position
	(*ChessMove)(nil),       // 6: rustchess.ChessMove
	(*AsyncRequest)(nil),    // 7: rustchess.AsyncRequest
	(*GameTime)(nil),        // 8: rustchess.GameTime
	(*GameStart)(nil),       // 9: rustchess.GameStart
}
var file_api_chessengine_proto_depIdxs = []int32{
	6,  // 0: rustchess.GameMessage.chess_move:type_name -> rustchess.ChessMove
	7,  // 1: rustchess.GameMessage.async_request:type_name -> rustchess.AsyncRequest
	9,  // 2: rustchess.GameMessage.game_start:type_name -> rustchess.GameStart
	0,  // 3: rustchess.PromotePiece.piece:type_name -> rustchess.PromotePiece.Piece
	5,  // 4: rustchess.ChessMove.original_position:type_name -> rustchess.Position
	5,  // 5: rustchess.ChessMove.new_position:type_name -> rustchess.Position
	4,  // 6: rustchess.ChessMove.promote_piece:type_name -> rustchess.PromotePiece
	1,  // 7: rustchess.AsyncRequest.type:type_name -> rustchess.AsyncRequest.Type
	2,  // 8: rustchess.GameStart.player_color:type_name -> rustchess.GameStart.Color
	8,  // 9: rustchess.GameStart.player_game_time:type_name -> rustchess.GameTime
	3,  // 10: rustchess.RustChess.Game:input_type -> rustchess.GameMessage
	3,  // 11: rustchess.RustChess.Game:output_type -> rustchess.GameMessage
	11, // [11:12] is the sub-list for method output_type
	10, // [10:11] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_api_chessengine_proto_init() }
func file_api_chessengine_proto_init() {
	if File_api_chessengine_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_chessengine_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PromotePiece); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Position); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChessMove); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AsyncRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameTime); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_chessengine_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameStart); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_api_chessengine_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*GameMessage_ChessMove)(nil),
		(*GameMessage_AsyncRequest)(nil),
		(*GameMessage_GameStart)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_chessengine_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_chessengine_proto_goTypes,
		DependencyIndexes: file_api_chessengine_proto_depIdxs,
		EnumInfos:         file_api_chessengine_proto_enumTypes,
		MessageInfos:      file_api_chessengine_proto_msgTypes,
	}.Build()
	File_api_chessengine_proto = out.File
	file_api_chessengine_proto_rawDesc = nil
	file_api_chessengine_proto_goTypes = nil
	file_api_chessengine_proto_depIdxs = nil
}
