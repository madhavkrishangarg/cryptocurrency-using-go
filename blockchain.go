package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt" // import the bolt package
	"log"
	"os"
)

const dbFile = "blockchain.db" // name of the database file
const blocksBucket = "blocks"  // name of the bucket
const genesisCoinbaseData = "03/04/2011 First Hosts To Win Cup, With Highest-Ever Runchase In Final"

type blockchain struct {
	tip []byte   // hash of the last block
	db  *bolt.DB // pointer to the database
}

func (bc *blockchain) mineBlock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.verifyTransaction(tx) != true {
			log.Panic("Error: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error { // read the last block hash from the database
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := newBlock(transactions, lastHash) // create a new block

	err = bc.db.Update(func(tx *bolt.Tx) error { // write the new block to the database
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *blockchain) findUnspentTransactions(pubKeyHash []byte) []Transaction { // find the unspent transactions

	var unspentTxs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.iterator()

	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs: // label the loop
			for outIdx, out := range tx.Vout { // iterate over the outputs
				if spentTXOs[txID] != nil { // check if the output is spent
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.isLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.isCoinbase() { // iterate over the inputs
				for _, in := range tx.Vin {
					if in.usesKey(pubKeyHash) { // check if the input uses the key
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break // break the loop if the genesis block is reached
		}
	}

	//print all unspent transactions
	// for _, tx := range unspentTxs {
	// 	log.Printf("Unspent transaction: %v", tx)
	// }

	// log.Printf("Total unspent transactions: %d", len(unspentTxs))

	return unspentTxs
}

func (bc *blockchain) findTransaction(ID []byte) (Transaction, error) { // find a transaction by its ID
	bci := bc.iterator()

	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break // break the loop if the genesis block is reached
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *blockchain) findUTXO(pubKeyHash []byte) []TXOutput {		// find the unspent transaction outputs
    var UTXOs []TXOutput
    unspentTransactions := bc.findUnspentTransactions(pubKeyHash)
    
    for _, tx := range unspentTransactions {
        for _, out := range tx.Vout {
            if out.isLockedWithKey(pubKeyHash) {
                UTXOs = append(UTXOs, out)
                // log.Printf("Found UTXO: Value=%d", out.Value)
            }
        }
    }
    
    // log.Printf("Total UTXOs found: %d", len(UTXOs))
    return UTXOs
}

func (bc *blockchain) findSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) { // find the spendable outputs
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.findUnspentTransactions(pubKeyHash)
	accumulated := 0

Find:
	for _, tx := range unspentTxs { // iterate over the unspent transactions
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout { // iterate over the outputs
			if out.isLockedWithKey(pubKeyHash) && accumulated < amount { // check if the output is locked with the key and the accumulated amount is less than the amount
				accumulated += out.Value                                    // add the output value to the accumulated amount
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx) // add the output index to the list of unspent outputs
				if accumulated >= amount { // check if the accumulated amount is greater than or equal to the amount
					break Find // break the loop
				}
			}
		}
	}
	// log.Printf("Total accumulated: %d, Required amount: %d", accumulated, amount)
    // log.Printf("Unspent outputs found: %v", unspentOutputs)

	return accumulated, unspentOutputs

}

func (bc *blockchain) iterator() *blockchainIterator {
	bci := &blockchainIterator{bc.tip, bc.db} // create a new iterator
	return bci
}

func dbExists() bool { // check if the database exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// func (bc *blockchain) addBlockUnsafe(data string){
// 	prevBlock := bc.blocks[len(bc.blocks)-1]
// 	newBlock := newBlockUnsafe(data, prevBlock.hash)
// 	bc.blocks = append(bc.blocks, newBlock)
// 	bc.hashes[string(newBlock.hash)] = newBlock
// }

func newBlockchain(address string) *blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one!")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil) // open the database
	if err != nil {                         // check for errors
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error { // write the genesis block to the database
		b := tx.Bucket([]byte(blocksBucket)) // get the bucket
		tip = b.Get([]byte("l"))             // get the last block hash

		return nil
	})

	if err != nil { // check for errors
		log.Panic(err)
	}

	bc := blockchain{tip, db} // create a new blockchain
	return &bc
}

func createBlockchain(address string) *blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	cbtx := newCoinbaseTX(address, genesisCoinbaseData) // create a coinbase transaction
	genesis := genesisBlock(cbtx)                       // create a genesis block

	db, err := bolt.Open(dbFile, 0600, nil) // open the database
	if err != nil {                         // check for errors
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error { // write the genesis block to the database
		b, err := tx.CreateBucket([]byte(blocksBucket)) // create a new bucket
		if err != nil {                                 // check for errors
			log.Panic(err)
		}
		err = b.Put(genesis.Hash, genesis.serialize()) // write the genesis block to the bucket
		if err != nil {                                // check for errors
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash) // update the last block hash
		if err != nil {                        // check for errors
			log.Panic(err)
		}
		tip = genesis.Hash // update the tip of the blockchain
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := blockchain{tip, db} // create a new blockchain

	return &bc
}

func (bc *blockchain) signTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.findTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.sign(privKey, prevTXs)
}

func (bc *blockchain) verifyTransaction(tx *Transaction) bool {
	if tx.isCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.findTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.verify(prevTXs)
}