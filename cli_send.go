package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from string, to string, amount int) {
	if !validateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}

	if !validateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := newBlockchain(from)
	defer bc.db.Close()

	tx := newUTXOTransaction(from, to, amount, bc)
	bc.mineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}