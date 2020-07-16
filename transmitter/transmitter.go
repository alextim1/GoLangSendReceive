// transmitter
// transmitter works as a Client. Tx sends generated message by RPC to the
// Server side (Receiver)
package main

import (
	"context"
	"errors"
	"fmt"
	"math"

	"time"
	tr "transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	HOST            = "127.0.0.1"
	PORT            = "9090"
	VALUESPERSECOND = 3
)

// Pipeline : Generator -> Retriever -> main (Send message)

func main() {

	// non Blocking channel
	messenger := make(chan tr.CarrierDto, math.MaxInt8)

	forSending := make(chan tr.CarrierDto)

	//Generator
	go func(messageCh chan tr.CarrierDto) {
		prevPrev := tr.CarrierDto{FibNumber: 0}
		prev := tr.CarrierDto{FibNumber: 1}
		messageCh <- prevPrev
		messageCh <- prev

		var err error

		for err == nil {
			var buffer int64
			buffer, err = Summator(prevPrev.FibNumber, prev.FibNumber)
			prevPrev.FibNumber = prev.FibNumber
			prev.FibNumber = buffer
			messageCh <- prev
		}

		fmt.Println("finish")
		closeReceiver := tr.CarrierDto{FibNumber: -1}
		messageCh <- closeReceiver
		fmt.Println("finish")

	}(messenger)

	//Retriever with asssigned speed for data stream
	go func(inputCh chan tr.CarrierDto, outputCh chan tr.CarrierDto) {
		message := <-inputCh
		for message.FibNumber != -1 {
			outputCh <- message
			time.Sleep(time.Second / VALUESPERSECOND)
			message = <-inputCh
		}

		closeReceiver := tr.CarrierDto{FibNumber: -1}
		outputCh <- closeReceiver

		fmt.Println("exit retriever")
	}(messenger, forSending)

	//main part - sending
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(HOST+":"+PORT, opts...)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	client := tr.NewSenderClient(conn)

	message := <-forSending
	for message.FibNumber != -1 {
		_, err := client.Send(context.Background(), &message)

		if err != nil {
			grpclog.Fatalf("fail to dial: %v", err)
		}
		message = <-forSending
		fmt.Println(message.FibNumber)
	}

	closeReceiver := tr.CarrierDto{FibNumber: -1}
	_, _ = client.Send(context.Background(), &closeReceiver)
	_, _ = client.Send(context.Background(), &closeReceiver)

	fmt.Println("exit App")
	return
}

func Summator(arg1 int64, arg2 int64) (int64, error) {
	if arg1 > math.MaxInt64-arg2 {
		return 0, errors.New("integer overflow")
	} else {
		return arg1 + arg2, nil
	}
}
