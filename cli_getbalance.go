package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !validateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := newBlockchain()
	UTXOSet := UTXOSet{bc}


	defer bc.db.Close()

	balance := 0
	pubKeyHash := base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.findUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}