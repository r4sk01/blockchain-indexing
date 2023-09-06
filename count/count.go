package main

import (
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

func unmarshalBlock(rawBlock []json.RawMessage) Block {
	var (
		header       Header
		transactions []Transaction
	)

	// Unmarshal header
	err := json.Unmarshal(rawBlock[0], &header)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	// Unmarshal transactions
	for i := 1; i < len(rawBlock); i++ {
		var tx Transaction
		err = json.Unmarshal(rawBlock[i], &tx)
		if err != nil {
			log.Fatalf("error while reading json file: %s", err)

		}
		transactions = append(transactions, tx)
	}

	return Block{Header: header, Transactions: transactions}
}

func parseFile(data []byte) Chain {
	var chain Chain

	var rawBlocks []json.RawMessage
	err := json.Unmarshal(data, &rawBlocks)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	for _, rawBlock := range rawBlocks {
		var block []json.RawMessage
		err = json.Unmarshal(rawBlock, &block)
		if err != nil {
			log.Fatalf("error while reading json file: %s", err)

		}
		chain = append(chain, unmarshalBlock(block))
	}

	return chain

}

func main() {

	jsonData, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	chain := parseFile(jsonData)

	fmt.Printf("Number of blocks: %d\n", len(chain))

	var sum int
	for _, block := range chain {
		sum += len(block.Transactions)
	}
	fmt.Printf("Number of transactions: %d\n", sum)

}
