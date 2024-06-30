package main

import "fmt"

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := newWallets(nodeID)
	address := wallets.createWallet()
	wallets.saveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}