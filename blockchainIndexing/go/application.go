package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
	rangeSize := flag.Int("r", 1, "size of key range")
	version := flag.Int("v", 1, "version to query for point query")
	start := flag.Int("start", 1, "start version for version query")
	end := flag.Int("end", 1, "end version for version query")
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
	case "getHistoryForAssetsOld":
		getHistoryForAssetsOld(contract, *key)
	case "getHistoryForAssetRangeOld":
		getHistoryForAssetRangeOld(contract, *key, *rangeSize)
	case "pointQueryOld":
		pointQueryOld(contract, *key, *version)
	case "versionQueryOld":
		versionQueryOld(contract, *key, *start, *end)

	// GetHistoryForKeys API Required
	case "getHistoryForAssets":
		getHistoryForAssets(contract, *key)
	case "getHistoryForAssetRange":
		getHistoryForAssetRange(contract, *key, *rangeSize)

	// GetVersionsForKey API Required
	case "pointQuery":
		pointQuery(contract, *key, *version)
	case "versionQuery":
		versionQuery(contract, *key, *start, *end)
	}

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

func BulkInvoke(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL must be provided and must be an absolute path")

	}

	jsonData, err := os.ReadFile(fileUrl)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	chain := parseFile(jsonData)

	startTime := time.Now()
	log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	// Insert N blocks at a time
	N := 1
	for i := 0; i < len(chain); i += N {
		chunkTime := time.Now()

		end := i + N
		if end > len(chain) {
			end = len(chain)
		}
		chunk := chain[i:end]

		var transactions []Transaction
		for _, block := range chunk {
			transactions = append(transactions, block.Transactions...)
		}

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
		log.Printf("Execution Time: %f sec at chunk %d", executionTime, i/N+1)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	log.Printf("Finished bulk transaction at time: %s\n", endTime.Format(time.UnixDate))
	log.Printf("Total execution time is: %f sec\n", executionTime)

}

func BulkInvokeParallel(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL is not absolute.")
	}

	jsonData, err := os.ReadFile(fileUrl)
	if err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup

	// Create a buffered channel to limit number of goroutines
	sem := make(chan bool, 10)

	chain := parseFile(jsonData)

	startTime := time.Now()
	log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	var chunkCounter int
	totalTx := 0

	CHUNK_LIMIT := 500
	var transactionChunk []Transaction
	for i := 0; i < len(chain); i++ {
		chunkTime := time.Now()

		currentBlock := chain[i]
		currentBlockNumTransactions := len(currentBlock.Transactions)

		if len(transactionChunk)+currentBlockNumTransactions < CHUNK_LIMIT {
			transactionChunk = append(transactionChunk, currentBlock.Transactions...)
			continue
		}

		totalTx += len(transactionChunk)

		chunkCounter++
		chunkBytes, err := json.Marshal(transactionChunk)
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
		log.Printf("Execution Time: %f sec at chunk %d with length %d. Cumulative total: %d\n", executionTime, chunkCounter, len(transactionChunk), totalTx)

		// Reset chunk to include only the current batch
		transactionChunk = append([]Transaction{}, currentBlock.Transactions...)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Drain the semaphore channel
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func Invoke(contract *gateway.Contract, fileUrl string) {
	log.Println("Submit individual orders")

	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		fmt.Println("File URL must be provided and must be an absolute path")
		os.Exit(1)
	}

	jsonData, err := os.ReadFile(fileUrl)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	chain := parseFile(jsonData)

	for i := 0; i < len(chain); i++ {
		transactions := chain[i].Transactions
		for _, transaction := range transactions {
			transactionBytes, err := json.Marshal(transaction)
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %s", err)
			}
			_, err = contract.SubmitTransaction("Create", string(transactionBytes))
			if err != nil {
				log.Fatalf("Failed to submit transaction: %s\n", err)
			}
		}
	}

	log.Println("Done")
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

	fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func getHistoryForAssetsOld(contract *gateway.Contract, keys string) {
	startTime := time.Now()

	keys_list := strings.Split(keys, ",")
	for _, key := range keys_list {
		result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		fmt.Println(string(result))
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func IncrementHex(s string) string {
	const HEX_TABLE = "0123456789abcdef"
	sPlusOne := make([]byte, len(s))
	carry := 1
	for i := len(s) - 1; i >= 2; i-- {
		digitVal := strings.IndexByte(HEX_TABLE, s[i])
		digitVal = digitVal + carry
		carry = digitVal / 16
		newDigitVal := digitVal % 16
		sPlusOne[i] = HEX_TABLE[newDigitVal]
	}
	return string(sPlusOne)
}

func getHistoryForAssetRangeOld(contract *gateway.Contract, key string, rangeSize int) {

	startTime := time.Now()

	numKeys := 0
	for i := 0; i < rangeSize; i++ {
		result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		if string(result) != "null" {
			numKeys++
		}
		// fmt.Println(string(result))
		key = IncrementHex(key)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	log.Printf("Total execution time is: %f sec\n", executionTime)
	log.Printf("%d keys found within rangeSize %d\n", numKeys, rangeSize)
}

func pointQueryOld(contract *gateway.Contract, key string, version int) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Timestamp < assets[j].Timestamp
	})

	if version < 0 || version > len(assets) {
		log.Fatalf("Version number out of range: %d\n", version)
	}

	selectedAsset := assets[version-1]

	assetJSON, err := json.Marshal(selectedAsset)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s\n", err)
	}
	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	fmt.Println(string(assetJSON))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

// versionQuery calls GetHistoryForKey API to execute Version Query
func versionQueryOld(contract *gateway.Contract, key string, start int, end int) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Timestamp < assets[j].Timestamp
	})

	if start < 1 || end < start || end > len(assets) {
		log.Fatalf("Start or end index out of range: start=%d, end=%d\n", start, end)
	}

	selectedAssets := assets[start-1 : end]

	endTime := time.Now()

	executionTime := endTime.Sub(startTime).Seconds()

	_, err = json.Marshal(selectedAssets)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s\n", err)
	}

	log.Printf("Total number of assets is: %d\n", len(assets))

	// fmt.Println(string(assetsJSON))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func getHistoryForAssets(contract *gateway.Contract, keys string) {
	startTime := time.Now()

	keys_list := strings.Split(keys, ",")
	result, err := contract.EvaluateTransaction("getHistoryForAssets", keys_list...)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func getHistoryForAssetRange(contract *gateway.Contract, key string, rangeSize int) {

	keys_list := []string{}
	for i := 0; i < rangeSize; i++ {
		keys_list = append(keys_list, key)
		key = IncrementHex(key)
	}

	startTime := time.Now()

	_, err := contract.EvaluateTransaction("getHistoryForAssets", keys_list...)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func pointQuery(contract *gateway.Contract, key string, version int) {

	fmt.Printf("Querying for version %d of key %s\n", version, key)
	startTime := time.Now()

	versionString := strconv.Itoa(version)

	result, err := contract.EvaluateTransaction("getVersionsForAsset", key, versionString, versionString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func versionQuery(contract *gateway.Contract, key string, start int, end int) {

	fmt.Printf("Querying for versions from %d to %d of key %s\n", start, end, key)
	startTime := time.Now()

	startString := strconv.Itoa(start)
	endString := strconv.Itoa(end)

	_, err := contract.EvaluateTransaction("getVersionsForAsset", key, startString, endString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
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
