// receiver works as an Server side
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	tr "transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	PORT          = "9090"
	LOGSPERSECOND = 2
)

// Pipeline : Receive -> Select Go to Logger or to the TRASH box

func main() {

	log := make(chan tr.CarrierDto)
	trash := make(chan tr.CarrierDto)

	listener, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	tr.RegisterSenderServer(grpcServer, &Server{loggerCh: log, trashBoxCh: trash})

	// Logging messages
	go func(log chan tr.CarrierDto) {
		msg := <-log

		fmt.Println("logged: %i", msg.FibNumber)

		for msg.FibNumber != -1 {
			time.Sleep(time.Second / LOGSPERSECOND)

			msg = <-log

			fmt.Println("logged: %i", msg.FibNumber)
		}

		fmt.Println("exit logger")

	}(log)

	go func(trash chan tr.CarrierDto) {
		msg := <-trash

		fmt.Println("trash: %i", msg.FibNumber)

		for msg.FibNumber != -1 {

			msg = <-trash

			fmt.Println("trash: %i", msg.FibNumber)
		}

		fmt.Println("exit trash box")

	}(trash)

	grpcServer.Serve(listener)
}

type Server struct {
	loggerCh   chan tr.CarrierDto
	trashBoxCh chan tr.CarrierDto
}

func (s *Server) Send(ctx context.Context,
	in *tr.CarrierDto,
) (*tr.Empthy, error) {
	select {
	case s.loggerCh <- *in:
	case s.trashBoxCh <- *in:
	default:
	}

	return &tr.Empthy{}, nil
}
