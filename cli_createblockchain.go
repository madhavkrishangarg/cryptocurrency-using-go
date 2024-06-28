package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !validateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := createBlockchain(address)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.reindex()

	fmt.Println("Done!")
}