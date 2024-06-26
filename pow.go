package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"math"
	"math/big"
)

const targetBits = 24		// the lower the targetBits, the more difficult it is to mine a block, represented by the number of leading zeroes in the hash

type proofOfWork struct {
	block  *block
	target *big.Int
}

func newPow(b *block) *proofOfWork { 			// create a new proof of work struct
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &proofOfWork{b, target}
}

func (p *proofOfWork) prepareData(nonce int) []byte {	// prepare the data to be hashed
	data := bytes.Join([][]byte{
		p.block.PrevBlockHash,
		p.block.hashTransactions(),
		[]byte(strconv.FormatInt(p.block.Timestamp, 10)),
		[]byte(strconv.FormatInt(int64(targetBits), 10)),
		[]byte(strconv.FormatInt(int64(nonce), 10)),
	}, []byte{})	// concatenate the byte slices, data = "prevBlockHash + data + timestamp + targetBits + nonce"
	return data
}

func (p *proofOfWork) run() (int, []byte) {	// run the proof of work algorithm
	var hashInt big.Int  // hashInt is a big.Int type to store the hash as an integer
	var hash [32]byte	// hash is a byte array of size 32
	nonce := 0			// nonce is the number of iterations of the proof of work algorithm

	fmt.Printf("Mining block")
	for nonce < math.MaxInt64 {	 
		data := p.prepareData(nonce) // prepare the data to be hashed
		hash = sha256.Sum256(data)			// hash the data
		hashInt.SetBytes(hash[:])			// set the hash as a big.Int type

		if hashInt.Cmp(p.target) == -1 {		// if the hash is less than the target, the block is mined
			fmt.Printf("Block mined - data: %s\n, hash: %x\n", data, hash)
			fmt.Printf("Nonce: %d\n", nonce)
			break
		} else {
			nonce++		// increment the nonce and try again	
		}
	}

	return nonce, hash[:] 	//return the nonce and hash
}

func (p *proofOfWork) validate() bool {
	var hashInt big.Int		// hashInt is a big.Int type to store the hash as an integer

	data := p.prepareData(p.block.Nonce)		
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(p.target) == -1
}