// Transmitter.
//
// 	Sample of the program which implements a Fibonacci numbers generator
// 	with preset frequency and send generated numbers to the Receiver side.
//
// 	Transmitter implements a Pipeline: Generator -> Retriever -> main (Send message).
// 	Transmitter works as a Client side which sends generated message by RPC to the
// 	Server side (Receiver).
// 	RPC has been implemented by using ProtoBuf protocol.
//
package main

import (
	"context"
	"errors"
	"fmt"
	"math"

	"time"
	tr "transport"

	"google.golang.org/grpc"
)

// Connection parameters.
const (
	HOST = "127.0.0.1"
	PORT = "9090"
)

// Parameter of creating messages frequency times per second.
const (
	ValuesPerSecond = 3
)

// Error raised when return value of function exceeds MaxInt64.
var (
	IntegerExceeding error = errors.New("integer overflow")
)

func main() {

	// Nonblocking channel which can contain all generated numbers before sending.
	messenger := make(chan tr.CarrierDto, math.MaxInt8)

	forSending := make(chan tr.CarrierDto)

	// Generator.
	// Generates Fibonacci numbers for further sending.
	go func(messageCh chan tr.CarrierDto) {
		prevPrev := tr.CarrierDto{FibNumber: 0}
		prev := tr.CarrierDto{FibNumber: 1}
		messageCh <- prevPrev
		messageCh <- prev
		closeReceiver := tr.CarrierDto{FibNumber: -1}

		defer fmt.Println("finish")

		var err error

		for err == nil {
			var buffer int64
			buffer, err = Adder(prevPrev.FibNumber, prev.FibNumber)
			prevPrev.FibNumber = prev.FibNumber
			prev.FibNumber = buffer
			messageCh <- prev
		}
		messageCh <- closeReceiver

	}(messenger)

	// Retriever. Retriever retrieves Fibonacci numbers with predefined speed for data stream.
	go func(inputCh chan tr.CarrierDto, outputCh chan tr.CarrierDto) {
		message := <-inputCh
		closeReceiver := tr.CarrierDto{FibNumber: -1}

		defer fmt.Println("exit retriever")

		for message.FibNumber != -1 {
			outputCh <- message
			time.Sleep(time.Second / ValuesPerSecond)
			message = <-inputCh
		}

		outputCh <- closeReceiver

	}(messenger, forSending)

	// Main part - sending messages to Receiver side.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(HOST+":"+PORT, opts...)

	if err != nil {
		fmt.Println("fail to dial: %v", err)
	}

	defer conn.Close()

	client := tr.NewSenderClient(conn)

	message := <-forSending
	for message.FibNumber != -1 {
		_, err := client.Send(context.Background(), &message)

		if err != nil {
			fmt.Println("fail to dial: %v", err)
		}
		message = <-forSending
		fmt.Println(message.FibNumber)
	}

	closeReceiver := tr.CarrierDto{FibNumber: -1}

	_, _ = client.Send(context.Background(), &closeReceiver)

	fmt.Println("exit App")
	return
}

// Adder function implements safety adding int64 numbers until
// the result does not exceed MaxInt64 value.
func Adder(arg1 int64, arg2 int64) (int64, error) {
	if arg1 > math.MaxInt64-arg2 {
		return 0, IntegerExceeding
	}
	return arg1 + arg2, nil
}
