package main

import "bytes"

type TXOutput struct {
	Value        int
	PubKeyHash   []byte
}

// Lock signs the output
func (out *TXOutput) lock(address []byte) {
	pubKeyHash := base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}



// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) isLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// NewTXOutput create a new TXOutput
func newTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.lock([]byte(address))

	return txo
}