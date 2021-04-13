package matchserver

import (
	"context"
	"github.com/Ekotlikoff/gochess/internal/model"
	"io"
	"log"
	"time"

	pb "github.com/Ekotlikoff/gochess/api"
	"google.golang.org/grpc"
)

func (matchingServer *MatchingServer) createEngineClient(
	engineAddr string, engineConnTimeout time.Duration) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(engineConnTimeout))
	conn, err := grpc.Dial(engineAddr, opts...)
	if err != nil {
		log.Println("ERROR: Failed to connect to chess engine at addr: " +
			engineAddr + " with error: " + err.Error())
	} else {
		log.Println("Succesfully connected to chess engine")
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
	go engineReceiveLoop(botPlayer, stream, waitc)
	for {
		select {
		case move := <-botPlayer.opponentPlayedMove:
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
	botPlayer *Player, stream pb.RustChess_GameClient, waitc chan struct{}) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			// read done.
			close(waitc)
			return
		}
		if err != nil {
			log.Printf("Failed to receive a msg : %v", err)
			// TODO resign in this case?
			close(waitc)
			return
		} else {
			botMove := pbToMove(in.GetChessMove())
			botPlayer.requestChanSync <- botMove
			<-botPlayer.responseChanSync
		}
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
