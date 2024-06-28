package main

import "bytes"

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}


func (in *TXInput) usesKey(pubKeyHash []byte) bool {
	lockingHash := hashPubKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}