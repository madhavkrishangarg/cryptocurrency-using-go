package main

import (
	"log"
	"github.com/boltdb/bolt" 			// import the bolt package
)

const dbFile = "blockchain.db"		// name of the database file
const blocksBucket = "blocks"		// name of the bucket

type blockchain struct {
	tip []byte		// hash of the last block
	db *bolt.DB		// pointer to the database
}

type blockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (bc *blockchain) addBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {		// read the last block hash
		b := tx.Bucket([]byte(blocksBucket))		// get the bucket
		lastHash = b.Get([]byte("l"))		// get the last block hash

		return nil		
	})

	if err != nil {		// check for errors
		log.Panic(err)
	}

	newBlock := newBlock(data, lastHash)		// create a new block
	err = bc.db.Update(func(tx *bolt.Tx) error {		// write the new block to the database
		b := tx.Bucket([]byte(blocksBucket))		// get the bucket
		err := b.Put(newBlock.hash, newBlock.serialize())		// write the block to the bucket
		if err != nil {		// check for errors
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.hash)		// update the last block hash
		if err != nil {		// check for errors
			log.Panic(err)
		}
		bc.tip = newBlock.hash		// update the tip of the blockchain
		return nil
	})
}

func (bc *blockchain) iterator() *blockchainIterator {
	bci := &blockchainIterator{bc.tip, bc.db}		// create a new iterator
	return bci
}

func (i *blockchainIterator) next() *block {
	var block *block

	err := i.db.View(func(tx *bolt.Tx) error {		// read the block from the database
		b := tx.Bucket([]byte(blocksBucket))		// get the bucket
		encodedBlock := b.Get(i.currentHash)		// get the block
		block = deserialize(encodedBlock)		// deserialize the block

		return nil
	})

	if err != nil {		// check for errors
		log.Panic(err)
	}

	i.currentHash = block.prevBlockHash		// update the current hash
	return block
}

// func (bc *blockchain) addBlockUnsafe(data string){
// 	prevBlock := bc.blocks[len(bc.blocks)-1]
// 	newBlock := newBlockUnsafe(data, prevBlock.hash)
// 	bc.blocks = append(bc.blocks, newBlock)
// 	bc.hashes[string(newBlock.hash)] = newBlock
// }

func newBlockchain() *blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)		// open the database
	if err != nil {		// check for errors
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {		// create the bucket
		b := tx.Bucket([]byte(blocksBucket))		// get the bucket
		if b == nil {		// if bucket does not exist
			genesis := genesisBlock()		// create the genesis block
			b, err := tx.CreateBucket([]byte(blocksBucket))		// create the bucket
			if err != nil {		// check for errors
				log.Panic(err)
			}
			err = b.Put(genesis.hash, genesis.serialize())		// write the genesis block to the bucket
			if err != nil {		// check for errors
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.hash)		// update the last block hash
			if err != nil {		// check for errors
				log.Panic(err)
			}
			tip = genesis.hash		// update the tip of the blockchain
		} else {
			tip = b.Get([]byte("l"))		// if bucket exists, get the last block hash
		}

		return nil
	})

	if err != nil {		// check for errors
		log.Panic(err)
	}

	bc := blockchain{tip, db}		// create the blockchain
	return &bc
}
