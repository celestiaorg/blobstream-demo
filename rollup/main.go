package main

import (
	"fmt"
)

type SpanSequence struct {
	height uint
	index  uint
	length uint
}

type RollupHeader struct {
	stateRoot []byte
	sequence  SpanSequence
}

type RollupTransaction struct {
	from   string
	to     string
	amount uint
}

func main() {
	// Create a new span sequence
	span := SpanSequence{
		height: 1,
		index:  0,
		length: 1,
	}

	// Create a new rollup header
	header := RollupHeader{
		stateRoot: []byte("stateRoot"),
		sequence:  span,
	}

	// Create a new rollup transaction
	transaction := RollupTransaction{
		from:   "0xFromAddress",
		to:     "0xToAddress",
		amount: 100,
	}

	// Print the created objects
	fmt.Println(header)
	fmt.Println(transaction)
}
