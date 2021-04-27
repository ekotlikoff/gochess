package matchserver

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"

	pb "github.com/Ekotlikoff/gochess/api"
	"google.golang.org/grpc"
)

func (matchingServer *MatchingServer) createEngineClient(
	engineAddr string, engineConnTimeout time.Duration) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	ctx, cancel := context.WithTimeout(context.Background(), engineConnTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, engineAddr, opts...)
	if err != nil {
		log.Println("ERROR: Failed to connect to chess engine at addr: " +
			engineAddr + " with error: " + err.Error())
	} else {
		log.Println("Successfully connected to chess engine")
		matchingServer.botMatchingEnabled = true
		matchingServer.engineClient = pb.NewRustChessClient(conn)
		matchingServer.engineClientConn = conn
	}
}

func (matchingServer *MatchingServer) engineSession(botPlayer *Player) {
	stream, err := matchingServer.engineClient.Game(context.Background())
	if err != nil {
		log.Println("FATAL: Failed to instantiate GRPC conn to engine")
	}
	err = botPlayer.WaitForMatchStart()
	if err != nil {
		log.Println("Bot failed to find match")
		return
	}
	botPlayer.SetSearchingForMatch(false)
	botPBColor := pb.GameStart_BLACK
	if botPlayer.Color() == model.White {
		botPBColor = pb.GameStart_WHITE
	}
	gameStartMsg := pb.GameMessage{
		Request: &pb.GameMessage_GameStart{
			GameStart: &pb.GameStart{
				PlayerColor: botPBColor,
				PlayerGameTime: &pb.GameTime{
					PlayerMainTime: uint32(botPlayer.MatchMaxTimeMs()),
				},
			},
		},
	}
	if err := stream.Send(&gameStartMsg); err != nil {
		log.Printf("FATAL: Failed to send gameStartMsg to bot: %v", err)
		// TODO resign in this case?
		return
	}
	gameOver := botPlayer.match.gameOver
	waitc := make(chan struct{})
	go engineReceiveLoop(matchingServer, botPlayer, stream, waitc)
	for {
		select {
		case move := <-botPlayer.OpponentPlayedMove:
			moveMsg := moveToPB(move)
			stream.Send(&moveMsg)
		case <-gameOver:
			stream.CloseSend()
			<-waitc
			botPlayer.ClientDoneWithMatch()
			return
		}
	}
}

func engineReceiveLoop(
	matchingServer *MatchingServer, botPlayer *Player,
	stream pb.RustChess_GameClient, waitc chan struct{}) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			close(waitc)
			return
		}
		if err != nil {
			log.Printf("Failed to receive a msg, closing engine conn: %v", err)
			if botPlayer.GetMatch() != nil && !botPlayer.GetMatch().GameOver() {
				botPlayer.RequestChanAsync <- RequestAsync{
					Resign: true,
				}
			}
			matchingServer.botMatchingEnabled = false
			close(waitc)
			return
		}
		botMove := pbToMove(in.GetChessMove())
		botPlayer.requestChanSync <- botMove
		<-botPlayer.ResponseChanSync
	}
}

func moveToPB(move model.MoveRequest) pb.GameMessage {
	return pb.GameMessage{
		Request: &pb.GameMessage_ChessMove{
			ChessMove: &pb.ChessMove{
				OriginalPosition: &pb.Position{
					File: uint32(move.Position.File),
					Rank: uint32(move.Position.Rank),
				},
				NewPosition: &pb.Position{
					File: uint32(int8(move.Position.File) + move.Move.X),
					Rank: uint32(int8(move.Position.Rank) + move.Move.Y),
				},
			},
		},
	}
}

func pbToMove(msg *pb.ChessMove) model.MoveRequest {
	return model.MoveRequest{
		Position: model.Position{
			File: uint8(msg.OriginalPosition.File),
			Rank: uint8(msg.OriginalPosition.Rank),
		},
		Move: model.Move{
			X: int8(msg.NewPosition.File -
				msg.OriginalPosition.File),
			Y: int8(msg.NewPosition.Rank -
				msg.OriginalPosition.Rank),
		},
	}

}
