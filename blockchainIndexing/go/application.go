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
	"regexp"
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
	key := flag.String("k", "", "key for GetHistoryForKey")
	rangeSize := flag.Int("r", 1, "size of key range")
	keys_file := flag.String("keylist", "./1M-versions.txt", "keys list for range query")
	version := flag.Int("v", 1, "version to query for point query")
	start := flag.Int("s", 1, "start version for version query or starting block for block range query")
	end := flag.Int("e", 1, "end version for version query or ending block for block range query")
	updates := flag.Int("u", 0, "minimum updates required in range to appear in results")
	pageSize := flag.Int("p", 1, "number of assets to fetch per chaincode call")
	flag.Parse()

	switch *transaction {
	case "BulkInvoke":
		BulkInvoke(contract, *file)
	case "BulkInvokeParallel":
		BulkInvokeParallel(contract, *file)
	case "Invoke":
		Invoke(contract, *file)
	case "GetState":
		GetState(contract, *key)
	case "GetHistoryForKey":
		GetHistoryForKey(contract, *key)
	case "GetHistoryForKeyRangeOld":
		GetHistoryForKeyRangeOld(contract, *key)
	case "GetHistoryForOrderedKeyRangeOld":
		GetHistoryForOrderedKeyRangeOld(contract, *key, *rangeSize, *keys_file)
	case "GetHistoryForVersionOld":
		GetHistoryForVersionOld(contract, *key, *version)
	case "GetHistoryForVersionRangeOld":
		GetHistoryForVersionRangeOld(contract, *key, *start, *end)

	// GetHistoryForKeyRange API Required
	case "GetHistoryForKeyRange":
		GetHistoryForKeyRange(contract, *key)
	case "GetHistoryForOrderedKeyRange":
		GetHistoryForOrderedKeyRange(contract, *key, *rangeSize, *keys_file)

	// GetHistoryForVersionRange API Required
	case "GetHistoryForVersion":
		GetHistoryForVersion(contract, *key, *version)
	case "GetHistoryForVersionRange":
		GetHistoryForVersionRange(contract, *key, *start, *end)
	case "GetHistoryForKeyPaginated":
		GetHistoryForKeyPaginated(contract, *key, *pageSize)

	// GetHistoryForBlockRange API Required
	case "GetHistoryForBlockRange":
		GetHistoryForBlockRange(contract, *start, *end, *updates)

	case "GetHistoryForVersionRangeFetchAll":
		GetHistoryForVersionRangeFetchAll(contract, *key, *pageSize, *start, *end)

	case "GetHistoryForVersionFetchAll":
		GetHistoryForVersionFetchAll(contract, *key, *pageSize, *version)
	}

}

func BulkInvoke(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL must be provided and must be an absolute path")

	}

	var totalTransactions int

	startTime := time.Now()
	// log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	file, err := os.Open(fileUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	var transactions []Transaction
	blockCounter := 1

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

		//blockTime := time.Now()
		blockBytes, err := json.Marshal(transactions)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %s", err)
		}

		_, err = contract.SubmitTransaction("CreateBulk", string(blockBytes))
		if err != nil {
			log.Fatalf("Failed to submit transaction: %s\n", err)
		}
		// executionTime := time.Since(blockTime).Seconds()
		// log.Printf("Execution Time: %f sec at block %d with length: %d\n", executionTime, blockCounter, len(transactions))
		blockCounter++
		totalTransactions += len(transactions)
		transactions = []Transaction{}

	}

	// Read the closing ']' of the outermost array
	if _, err := decoder.Token(); err != nil {
		log.Fatal(err)
	}

	executionTime := time.Since(startTime).Seconds()
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
	// log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	file, err := os.Open(fileUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(bufio.NewReader(file))

	var transactions []Transaction
	blockCounter := 1

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

		// blockTime := time.Now()
		blockBytes, err := json.Marshal(transactions)
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
		}(string(blockBytes))
		// endTime := time.Now()
		// executionTime := endTime.Sub(blockTime).Seconds()
		// log.Printf("Execution Time: %f sec at block %d with length: %d\n", executionTime, blockCounter, len(transactions))
		blockCounter++
		totalTransactions += len(transactions)
		transactions = []Transaction{}

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

	executionTime := time.Since(startTime).Seconds()
	log.Printf("Time to insert %d transactions: %f seconds\n", totalTransactions, executionTime)
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

func GetState(contract *gateway.Contract, key string) {
	log.Println("-----stub.GetState() Test-----")

	result, err := contract.EvaluateTransaction("GetState", key)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %s\n", err)
	}
	tx := Transaction{}

	json.Unmarshal(result, &tx)
	log.Printf("%+v\n", tx)
}

// GetHistoryForKey calls GetHistoryForKey API
func GetHistoryForKey(contract *gateway.Contract, key string) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("GetHistoryForKey", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}
	fmt.Printf("Number of records found: %d\n", len(assets))

	//fmt.Printf("%+v\n", assets[0])
	log.Printf("Total execution time is: %f sec\n", executionTime)
	//index_total, index_average, disk_total, disk_average := get_read_times()
	//log.Printf("Total time to read index is %d microseconds with average time of %f microseconds\n", index_total, index_average)
	//log.Printf("Total time to read disk is %d microseconds with average time of %f microseconds\n", disk_total, disk_average)
}

func GetHistoryForKeyRangeOld(contract *gateway.Contract, keys string) {
	startTime := time.Now()

	keys_list := strings.Split(keys, ",")
	for _, key := range keys_list {
		result, err := contract.EvaluateTransaction("GetHistoryForKey", key)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		fmt.Println(string(result))
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Total execution time is: %f sec\n", executionTime)
}

// func IncrementHex(s string) string {
// 	const HEX_TABLE = "0123456789abcdef"
// 	sPlusOne := make([]byte, len(s))
// 	carry := 1
// 	for i := len(s) - 1; i >= 2; i-- {
// 		digitVal := strings.IndexByte(HEX_TABLE, s[i])
// 		digitVal = digitVal + carry
// 		carry = digitVal / 16
// 		newDigitVal := digitVal % 16
// 		sPlusOne[i] = HEX_TABLE[newDigitVal]
// 	}
// 	return string(sPlusOne)
// }

func GetHistoryForOrderedKeyRangeOld(contract *gateway.Contract, key string, rangeSize int, keys_file string) {
	all_keys := []string{}
	file, err := os.Open(keys_file)
	if err != nil {
		log.Fatalf("Failed to open keys file: %s\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		next_key := strings.Split(scanner.Text(), " ")[0]
		all_keys = append(all_keys, next_key)
	}

	var key_index int

	for key_index < len(all_keys) && all_keys[key_index] != key {
		key_index++
	}

	if key_index >= len(all_keys) {
		log.Fatalf("Key %s not found\n", key)
	}

	end := key_index + rangeSize
	if key_index+rangeSize >= len(all_keys) {
		end = len(all_keys) - 1
	}
	keys_list := all_keys[key_index:end]

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %s\n", err)
		return
	}

	startTime := time.Now()

	for _, key := range keys_list {
		_, err = contract.EvaluateTransaction("GetHistoryForKey", key)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}
		// fmt.Println(string(result))
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForVersionOld(contract *gateway.Contract, key string, version int) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("GetHistoryForKey", key)
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
	fmt.Printf("Retrieved %d total assets\n", len(assets))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

// GetHistoryForVersionRange calls GetHistoryForKey API to execute Version Query
func GetHistoryForVersionRangeOld(contract *gateway.Contract, key string, start int, end int) {
	startTime := time.Now()

	result, err := contract.EvaluateTransaction("GetHistoryForKey", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

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

	_, err = json.Marshal(selectedAssets)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s\n", err)
	}

	log.Printf("Total number of assets is: %d\n", len(assets))

	// fmt.Println(string(assetsJSON))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForKeyRange(contract *gateway.Contract, keys string) {
	startTime := time.Now()

	keys_list := strings.Split(keys, ",")
	_, err := contract.EvaluateTransaction("GetHistoryForKeyRange", keys_list...)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	// fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForOrderedKeyRange(contract *gateway.Contract, key string, rangeSize int, keys_file string) {

	all_keys := []string{}
	file, err := os.Open(keys_file)
	if err != nil {
		log.Fatalf("Failed to open keys file: %s\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		next_key := strings.Split(scanner.Text(), " ")[0]
		all_keys = append(all_keys, next_key)
	}

	var key_index int

	for key_index < len(all_keys) && all_keys[key_index] != key {
		key_index++
	}

	if key_index >= len(all_keys) {
		log.Fatalf("Key %s not found\n", key)
	}

	end := key_index + rangeSize
	if key_index+rangeSize >= len(all_keys) {
		end = len(all_keys) - 1
	}
	keys_list := all_keys[key_index:end]

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %s\n", err)
		return
	}

	startTime := time.Now()

	result, err := contract.EvaluateTransaction("GetHistoryForKeyRange", keys_list...)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}

	log.Printf("Total number of assets is: %d\n", len(assets))
	// fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForVersion(contract *gateway.Contract, key string, version int) {

	fmt.Printf("Querying for version %d of key %s\n", version, key)
	startTime := time.Now()

	versionString := strconv.Itoa(version)

	result, err := contract.EvaluateTransaction("GetHistoryForVersionRange", key, versionString, versionString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForVersionRange(contract *gateway.Contract, key string, start int, end int) {

	fmt.Printf("Querying for versions from %d to %d of key %s\n", start, end, key)
	startTime := time.Now()

	startString := strconv.Itoa(start)
	endString := strconv.Itoa(end)

	result, err := contract.EvaluateTransaction("GetHistoryForVersionRange", key, startString, endString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}

	log.Printf("Total number of assets is: %d\n", len(assets))
	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForBlockRange(contract *gateway.Contract, start int, end int, updates int) {

	fmt.Printf("Querying for keys within block range from %d to %d updated %d times or more \n", start, end, updates)
	startTime := time.Now()

	startString := strconv.Itoa(start)
	endString := strconv.Itoa(end)
	updatesString := strconv.Itoa(updates)

	result, err := contract.EvaluateTransaction("GetHistoryForBlockRange", startString, endString, updatesString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s\n", err)
	}

	log.Printf("Total number of assets is: %d\n", len(assets))
	// fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForKeyPaginated(contract *gateway.Contract, key string, pageSize int) {

	fmt.Printf("Fetching history for key %s with %d results at a time\n", key, pageSize)
	startTime := time.Now()

	var totalAssets []Asset
	start := 1
	end := pageSize
	for {
		fmt.Printf("Calling GetHistoryForVersionRange with start %d and end %d\n", start, end)
		result, err := contract.EvaluateTransaction("GetHistoryForVersionRange", key, strconv.Itoa(start), strconv.Itoa(end))
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}

		var currentAssets []Asset
		err = json.Unmarshal(result, &currentAssets)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %s\n", err)
		}
		totalAssets = append(totalAssets, currentAssets...)
		if len(currentAssets) < pageSize {
			break
		}
		start = end + 1
		end = start + pageSize - 1
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Total number of assets is: %d\n", len(totalAssets))
	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForVersionRangeFetchAll(contract *gateway.Contract, key string, pageSize int, start int, end int) {

	fmt.Printf("Fetching history for key %s with %d results at a time\n", key, pageSize)
	startTime := time.Now()

	var totalAssets []Asset
	pageStart := 1
	pageEnd := pageSize
	for {
		fmt.Printf("Calling GetHistoryForVersionRange with start %d and end %d\n", pageStart, pageEnd)
		result, err := contract.EvaluateTransaction("GetHistoryForVersionRange", key, strconv.Itoa(pageStart), strconv.Itoa(pageEnd))
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}

		var currentAssets []Asset
		err = json.Unmarshal(result, &currentAssets)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %s\n", err)
		}
		totalAssets = append(totalAssets, currentAssets...)
		if len(currentAssets) < pageSize {
			break
		}
		pageStart = pageEnd + 1
		pageEnd = pageStart + pageSize - 1
	}

	requestedAssets := totalAssets[start : end+1]

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Total number of assets is: %d\n", len(totalAssets))
	log.Printf("Requested assets found: %d\n", len(requestedAssets))
	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func GetHistoryForVersionFetchAll(contract *gateway.Contract, key string, pageSize int, version int) {

	fmt.Printf("Fetching history for key %s with %d results at a time\n", key, pageSize)
	startTime := time.Now()

	var totalAssets []Asset
	pageStart := 1
	pageEnd := pageSize
	for {
		fmt.Printf("Calling GetHistoryForVersionRange with start %d and end %d\n", pageStart, pageEnd)
		result, err := contract.EvaluateTransaction("GetHistoryForVersionRange", key, strconv.Itoa(pageStart), strconv.Itoa(pageEnd))
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %s\n", err)
		}

		var currentAssets []Asset
		err = json.Unmarshal(result, &currentAssets)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %s\n", err)
		}
		totalAssets = append(totalAssets, currentAssets...)
		if len(currentAssets) < pageSize {
			break
		}
		pageStart = pageEnd + 1
		pageEnd = pageStart + pageSize - 1
	}

	requestedAssets := totalAssets[version-1]

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	log.Printf("Total number of assets is: %d\n", len(totalAssets))
	log.Printf("Requested assets found: %d\n", requestedAssets)
	//fmt.Println(string(result))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func calculate_total_and_average(arr []int) (int, float64) {
	sum := 0
	for _, value := range arr {
		sum += value
	}

	if len(arr) == 0 {
		return 0, 0.0
	}

	return sum, float64(sum) / float64(len(arr))
}

func get_read_times() (int, float64, int, float64) {
	time_file, err := os.Open("/home/andrey/Documents/insert-tpch/blockchain-indexing/test-network/peerStorage2/read_times.txt")
	if err != nil {
		log.Printf("ERROR: Could not open time file: %s\n", err)
	}
	defer time_file.Close()

	scanner := bufio.NewScanner(time_file)
	index_times := []int{}
	disk_times := []int{}

	for scanner.Scan() {
		line := scanner.Text()
		index_pattern := `Time to read index: (\d+)`
		disk_pattern := `Time to read disk: (\d+)`
		re_index := regexp.MustCompile(index_pattern)
		re_disk := regexp.MustCompile(disk_pattern)

		if re_index.MatchString(line) {
			matches := re_index.FindStringSubmatch(line)
			value, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Fatalf("Error converting time: %s\n", err)
			}
			index_times = append(index_times, value)
		} else if re_disk.MatchString(line) {
			matches := re_disk.FindStringSubmatch(line)
			value, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Fatalf("Error converting time: %s\n", err)
			}
			disk_times = append(disk_times, value)
		}
	}
	index_total, index_average := calculate_total_and_average(index_times)
	disk_total, disk_average := calculate_total_and_average(disk_times)

	return index_total, index_average, disk_total, disk_average
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
