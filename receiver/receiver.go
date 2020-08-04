// Receiver.
//	Receiver listen a messages which are sent from the Transmitter side and loggs them to the file.
//	Logger works with a predefined frequency which is lower than frequency of sending messages.
//	When the Logger is not ready to logg message, message is thrown to the trash box (ignored).
//
//	Receiver implements next Pipeline : Receive -> Select Go to Logger or to the TRASH box.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	tr "transport"

	"google.golang.org/grpc"
)

// Connection parameters.
const (
	PORT = "9090"
)

// Log file name
const (
	LOGFILE = "log.txt"
)

// Logger frequency in logs per second.
const (
	LogsPerSecond = 2
)

func main() {

	log := make(chan tr.CarrierDto)
	trash := make(chan tr.CarrierDto)
	wg := &sync.WaitGroup{}
	ctx, finish := context.WithCancel(context.Background())

	listener, err := net.Listen("tcp", ":"+PORT)

	if err != nil {
		fmt.Println("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	tr.RegisterSenderServer(grpcServer, &Server{loggerCh: log, trashBoxCh: trash})

	// Logging messages.
	wg.Add(1)
	go func(ctx context.Context, log chan tr.CarrierDto) {
		path, _ := os.Getwd()
		file, err := os.Create(path + "/" + LOGFILE)

		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		defer wg.Done()
		defer fmt.Println("exit logger")
		defer file.Close()

		msg := <-log

		fmt.Println("logged: %i", msg.FibNumber)
		file.WriteString(fmt.Sprintf("%d\n", msg.FibNumber))

		for msg.FibNumber != -1 {
			time.Sleep(time.Second / LogsPerSecond)

			select {
			case msg = <-log:
			case <-ctx.Done():
				return
			}

			fmt.Println("logged: %i", msg.FibNumber)
			file.WriteString(fmt.Sprintf("%d\n", msg.FibNumber))
		}
		finish()

	}(ctx, log)

	// Goroutine throws redundant messages to the trash box when logger are not
	// ready to log them.
	wg.Add(1)
	go func(ctx context.Context, trash chan tr.CarrierDto) {
		defer wg.Done()
		defer fmt.Println("exit trash box")

		msg := <-trash

		fmt.Println("trash: %i", msg.FibNumber)

		for msg.FibNumber != -1 {
			select {
			case msg = <-trash:
			case <-ctx.Done():
				return
			}

			fmt.Println("trash: %i", msg.FibNumber)
		}

		finish()

	}(ctx, trash)

	// Waiting until goroutines which sort messages between logger and trash box
	// receive message with -1 value (stop sign).
	go func() {
		wg.Wait()
		grpcServer.Stop()
	}()
	grpcServer.Serve(listener)
	fmt.Println("exit App")
}

// Server structure represent a data channels which are filled during Server listening.
type Server struct {
	loggerCh   chan tr.CarrierDto
	trashBoxCh chan tr.CarrierDto
}

// Send method of server which is called remoutly by the Transmitter side.
func (s *Server) Send(ctx context.Context,
	in *tr.CarrierDto,
) (*tr.Empthy, error) {
	select {
	case s.loggerCh <- *in:
	case s.trashBoxCh <- *in:
	}
	return &tr.Empthy{}, nil
}
