package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from string, to string, amount int, nodeID string, mineNow bool) {
	if !validateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}

	if !validateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := newBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := newWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.getWallet(from)

	tx := newUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := newCoinbaseTX(from, "")
		txs := []*Transaction{cbTx, tx}

		newBlock := bc.mineBlock(txs)
		UTXOSet.update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}