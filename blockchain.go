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

const dbFile = "blockchain_%s.db" // name of the database file
const blocksBucket = "blocks"  // name of the bucket
const genesisCoinbaseData = "03/04/2011 First Hosts To Win Cup, With Highest-Ever Runchase In Final"

type blockchain struct {
	tip []byte   // hash of the last block
	db  *bolt.DB // pointer to the database
}

func (bc *blockchain) mineBlock(transactions []*Transaction) *block { // mine a new block
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		//ignore transactions that are not valid
		if bc.verifyTransaction(tx) == false {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error { // read the last block hash from the database
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := deserialize(blockData)
		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := newBlock(transactions, lastHash, lastHeight + 1) // create a new block

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

	return newBlock

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

func (bc *blockchain) findUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.iterator()

	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.isCoinbase() {
				for _, in := range tx.Vin {
					inTXID := hex.EncodeToString(in.Txid)
					spentTXOs[inTXID] = append(spentTXOs[inTXID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}

	return UTXO
}

func (bc *blockchain) iterator() *blockchainIterator {
	bci := &blockchainIterator{bc.tip, bc.db} // create a new iterator
	return bci
}

func dbExists(dbFile string) bool { // check if the database exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func newBlockchain(nodeID string) *blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)

	if dbExists(dbFile) == false { // check if the database exists
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

func createBlockchain(address string, nodeID string) *blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	
	if dbExists(dbFile) {
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

func (bc *blockchain) addBlock(block *block) { // add a block to the blockchain
	err := bc.db.Update(func(tx *bolt.Tx) error { // write the block to the database
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(block.Hash)

		if blockData != nil {
			return nil
		}

		err := b.Put(block.Hash, block.serialize())
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *blockchain) getBestHeight() int { // get the height of the last block
	var lastBlock *block

	err := bc.db.View(func(tx *bolt.Tx) error { // read the last block from the database
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = deserialize(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

func (bc *blockchain) getBlock(hash []byte) (*block, error) { // get a block by its hash
	var block *block

	err := bc.db.View(func(tx *bolt.Tx) error { // read the block from the database
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(hash)

		if blockData == nil {
			return errors.New("Block is not found")
		}

		block = deserialize(blockData)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (bc *blockchain) getBlockHashes() [][]byte { // get the hashes of all blocks in the blockchain
	var blocks [][]byte

	bci := bc.iterator()

	for {
		block := bci.next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
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
