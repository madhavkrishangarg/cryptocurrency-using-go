package main

import (
	"fmt"
)

func (cli* CLI) reindexUTXO(nodeID string) {
	bc := newBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	UTXOSet.reindex()

	count := UTXOSet.countTransactions()
	fmt.Printf("Reindexed UTXO set with %d transactions.\n", count)
}