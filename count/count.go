package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Chain []Block

type Block struct {
	Header       Header
	Transactions []Transaction
}

type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

type Header struct {
	BaseFeePerGas    string   `json:"baseFeePerGas"`
	Difficulty       string   `json:"difficulty"`
	ExtraData        string   `json:"extraData"`
	GasLimit         int      `json:"gasLimit"`
	GasUsed          int      `json:"gasUsed"`
	Hash             string   `json:"hash"`
	LogsBloom        string   `json:"logsBloom"`
	Miner            string   `json:"miner"`
	MixHash          string   `json:"mixHash"`
	Nonce            string   `json:"nonce"`
	Number           int      `json:"blockHash"`
	ParentHash       string   `json:"parentHash"`
	ReceiptsRoot     string   `json:"receiptsRoot"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	Size             int      `json:"size"`
	StateRoot        string   `json:"stateRoot"`
	Timestamp        int      `json:"timestamp"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	Transactions     []string `json:"transactions"`
	TransactionsRoot string   `json:"transactionsRoot"`
	Uncles           []string `json:"uncles"`
}

type Transaction struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber int    `json:"blockNumber"`
	From        string `json:"from"`
	Gas         int    `json:"gas"`
	GasPrice    string `json:"gasPrice"`

	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`

	Hash             string            `json:"hash"`
	Input            string            `json:"input"`
	Nonce            int               `json:"nonce"`
	To               string            `json:"to"`
	TransactionIndex int               `json:"transactionIndex"`
	Value            string            `json:"value"`
	Type             string            `json:"type"`
	AccessList       []AccessListEntry `json:"accessList"`
	ChainId          string            `json:"chainId"`
	V                string            `json:"v"`
	R                string            `json:"r"`
	S                string            `json:"s"`
}

func main() {

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	var transactions []Transaction

	// Read the opening '['
	if _, err := decoder.Token(); err != nil {
		panic(err)
	}

	// Iterate over blocks
	for decoder.More() {
		// Read the opening '[' of the block
		if _, err := decoder.Token(); err != nil {
			panic(err)
		}

		// Process the block header
		var blockHeader Header
		if err := decoder.Decode(&blockHeader); err != nil {
			panic(err)
		}

		// Process transactions
		for decoder.More() {
			var transaction Transaction
			if err := decoder.Decode(&transaction); err != nil {
				panic(err)
			}
			transactions = append(transactions, transaction)
		}

		// Read the closing ']' of the block
		if _, err := decoder.Token(); err != nil {
			panic(err)
		}
	}

	// Read the closing ']' of the outermost array
	if _, err := decoder.Token(); err != nil {
		panic(err)
	}

	fmt.Printf("Number of transactions: %d\n", len(transactions))

}
