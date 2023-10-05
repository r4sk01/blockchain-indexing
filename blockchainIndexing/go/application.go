package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Asset struct {
	Key       string                 `json:"key"`
	Record    map[string]interface{} `json:"record"`
	Timestamp string                 `json:"timestamp"`
}

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

	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %s\n", err)

	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %s\n", err)

	}

	contract := network.GetContract("blockchainIndexing")

	transaction := flag.String("t", "defaultQuery", "Choose a transaction to run")
	file := flag.String("f", "~", "file path for json data")
	key := flag.String("k", "", "key for getHistoryForAsset")
	version := flag.String("v", "1", "version to query for point query")
	startV := flag.String("startV", "1", "start version for version query")
	endV := flag.String("endV", "1", "end version for version query")
	startK := flag.String("startK", "1", "start key")
	endK := flag.String("endK", "1", "end key")
	startB := flag.String("startB", "1", "start block")
	endB := flag.String("endB", "1", "end block")
	flag.Parse()

	// /var/hyperledger/production/ledgersData/historyLeveldb

	switch *transaction {
	case "BulkInvoke":
		BulkInvoke(contract, *file)
	case "BulkInvokeParallel":
		BulkInvokeParallel(contract, *file)
	case "Invoke":
		Invoke(contract, *file)
	case "getHistoryForAsset":
		getHistoryForAsset(contract, *key)
	case "histTest":
		histTest(contract, *startK, *endK, *startB, *endB)
	case "pointQuery":
		pointQuery(contract, *key, *version, *startB, *endB)
	case "versionQuery":
		versionQuery(contract, *key, *startV, *endV, *startB, *endB)
	case "rangeQuery":
		rangeQuery(contract, *startK, *endK, *startB, *endB)
	case "getState":
		getState(contract, *key)
	}
}

func BulkInvoke(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL must be provided and must be an absolute path")

	}

	// Insert N transactions at a time
	var totalTransactions int

	startTime := time.Now()
	log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	file, err := os.Open(fileUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	var transactions []Transaction
	chunkCounter := 1

	// Read the opening '['
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	// Iterate over blocks
	for decoder.More() {
		// Read the opening '[' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}

		// Process the block header
		var blockHeader Header
		if err := decoder.Decode(&blockHeader); err != nil {
			log.Fatal(err)
		}

		// Process transactions
		for decoder.More() {
			var transaction Transaction
			if err := decoder.Decode(&transaction); err != nil {
				log.Fatal(err)
			}
			transactions = append(transactions, transaction)
		}

		// Read the closing ']' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}

		if len(transactions) > 0 {
			chunkTime := time.Now()
			chunkBytes, err := json.Marshal(transactions)
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %s", err)
			}

			_, err = contract.SubmitTransaction("CreateBulk", string(chunkBytes))
			if err != nil {
				log.Fatalf("Failed to submit transaction: %s\n", err)
			}
			endTime := time.Now()
			executionTime := endTime.Sub(chunkTime).Seconds()
			log.Printf("Execution Time: %f sec at chunk %d with length: %d\n", executionTime, chunkCounter, len(transactions))
			chunkCounter++
			totalTransactions += len(transactions)
			transactions = []Transaction{}
		}
	}

	// Read the closing ']' of the outermost array
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	log.Printf("Finished bulk transaction at time: %s\n", endTime.Format(time.UnixDate))
	log.Printf("Total execution time is: %f sec\n", executionTime)
	log.Printf("Total of %d transactions inserted\n", totalTransactions)

}

func BulkInvokeParallel(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL is not absolute.")
	}
	var totalTransactions int

	var wg sync.WaitGroup

	// Create a buffered channel to limit number of goroutines
	sem := make(chan bool, 10)

	startTime := time.Now()
	log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	file, err := os.Open(fileUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	var transactions []Transaction
	chunkCounter := 1

	// Read the opening '['
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	// Iterate over blocks
	for decoder.More() {
		// Read the opening '[' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}

		// Process the block header
		var blockHeader Header
		if err := decoder.Decode(&blockHeader); err != nil {
			log.Fatal(err)
		}

		// Process transactions
		for decoder.More() {
			var transaction Transaction
			if err := decoder.Decode(&transaction); err != nil {
				log.Fatal(err)
			}
			transactions = append(transactions, transaction)
		}

		// Read the closing ']' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}

		if len(transactions) > 0 {
			chunkTime := time.Now()
			chunkBytes, err := json.Marshal(transactions)
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %s", err)
			}
			wg.Add(1)
			// Before spawning a goroutine, acquire a slot in the channel
			sem <- true
			go func(data string) {
				defer wg.Done()
				_, err = contract.SubmitTransaction("CreateBulkParallel", data)
				if err != nil {
					log.Println(err)
				}
				// Once the transaction is complete, release the slot
				<-sem
			}(string(chunkBytes))
			endTime := time.Now()
			executionTime := endTime.Sub(chunkTime).Seconds()
			log.Printf("Execution Time: %f sec at chunk %d with length: %d\n", executionTime, chunkCounter, len(transactions))
			chunkCounter++
			totalTransactions += len(transactions)
			transactions = []Transaction{}
		}

	}

	// Read the closing ']' of the outermost array
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Drain the semaphore channel
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	log.Printf("Total of %d transactions inserted\n", totalTransactions)
}

func Invoke(contract *gateway.Contract, fileUrl string) {
	log.Println("Submit individual orders")

	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		fmt.Println("File URL must be provided and must be an absolute path")
		os.Exit(1)
	}

	var totalTransactions int

	file, err := os.Open(fileUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	// Read the opening '['
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	// Iterate over blocks
	for decoder.More() {
		// Read the opening '[' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}

		// Process the block header
		var blockHeader Header
		if err := decoder.Decode(&blockHeader); err != nil {
			log.Fatal(err)
		}

		// Process transactions
		for decoder.More() {
			txStart := time.Now()
			var transaction Transaction
			if err := decoder.Decode(&transaction); err != nil {
				log.Fatal(err)
			}
			transactionBytes, err := json.Marshal(transaction)
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %s", err)
			}

			_, err = contract.SubmitTransaction("Create", string(transactionBytes))
			if err != nil {
				log.Fatalf("Failed to submit transaction: %s\n", err)
			}
			endTime := time.Now()
			executionTime := endTime.Sub(txStart).Seconds()
			log.Printf("Execution Time: %f sec\n", executionTime)
			totalTransactions++
		}

		// Read the closing ']' of the block
		if _, err := decoder.Token(); err != nil {
			log.Fatal(err)
		}
	}

	// Read the closing ']' of the outermost array
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Total of %d transactions inserted\n", totalTransactions)
}

// getHistoryForAsset calls GetHistoryForKey API
func getHistoryForAsset(contract *gateway.Contract, key string) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	fmt.Printf("%s\n", result)
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func histTest(contract *gateway.Contract, startKey string, endKey string, startBlk string, endBlk string) {
	log.Println("-----stub.Hist() Test-----")

	_, err := contract.EvaluateTransaction("histTest", startKey, endKey, startBlk, endBlk)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %s\n", err)
	}
	log.Println("Transaction has been evaluated")
}

func pointQuery(contract *gateway.Contract, key string, version string, startBlk string, endBlk string) {
	log.Println("-----Point Query-----")
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("PointQuery", key, version, startBlk, endBlk)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Transaction has been evaluated, result is: %s\n", string(result))
	log.Printf("Finished point query with execution time: %f sec\n", executionTime)
}

func versionQuery(contract *gateway.Contract, key string, startVersion string, endVersion string, startBlk string, endBlk string) {
	log.Println("-----Version Query-----")
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("VersionQuery", key, startVersion, endVersion, startBlk, endBlk)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}
	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	log.Printf("Transaction has been evaluated, result is: %s\n", string(result))

	log.Printf("Finished point query with execution time: %f sec\n", executionTime)
}

func rangeQuery(contract *gateway.Contract, startKey string, endKey string, startBlk string, endBlk string) {
	log.Println("-----Range Query-----")
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("RangeQuery", startKey, endKey, startBlk, endBlk)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Transaction has been evaluated, result is: %s\n", string(result))

	log.Printf("Finished point query with execution time: %f sec\n", executionTime)
}

func getState(contract *gateway.Contract, key string) {
	log.Println("-----stub.GetState() Test-----")

	result, err := contract.EvaluateTransaction("getState", key)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %s\n", err)
	}
	tx := Transaction{}

	json.Unmarshal(result, &tx)
	log.Printf("+%v\n", tx)
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}
