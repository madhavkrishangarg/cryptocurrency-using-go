package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := newWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.getAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}