package main

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func newMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := MerkleNode{nil, nil, datum}
		nodes = append(nodes, node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := MerkleNode{&nodes[j], &nodes[j+1], nil}
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

func newMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{left, right, data}

	if left != nil && right != nil {
		hash := sha256.Sum256(append(left.Data, right.Data...))
		node.Data = hash[:]
	}

	return &node
}