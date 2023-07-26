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

type Table struct {
	Table []Order `json:"table"`
}

type Order struct {
	L_ORDERKEY      int     `json:"L_ORDERKEY"`
	L_PARTKEY       int     `json:"L_PARTKEY"`
	L_SUPPKEY       int     `json:"L_SUPPKEY"`
	L_LINENUMBER    int     `json:"L_LINENUMBER"`
	L_QUANTITY      int     `json:"L_QUANTITY"`
	L_EXTENDEDPRICE float64 `json:"L_EXTENDEDPRICE"`
	L_DISCOUNT      float64 `json:"L_DISCOUNT"`
	L_TAX           float64 `json:"L_TAX"`
	L_RETURNFLAG    string  `json:"L_RETURNFLAG"`
	L_LINESTATUS    string  `json:"L_LINESTATUS"`
	L_SHIPDATE      string  `json:"L_SHIPDATE"`
	L_COMMITDATE    string  `json:"L_COMMITDATE"`
	L_RECEIPTDATE   string  `json:"L_RECEIPTDATE"`
	L_SHIPINSTRUCT  string  `json:"L_SHIPINSTRUCT"`
	L_SHIPMODE      string  `json:"L_SHIPMODE"`
	L_COMMENT       string  `json:"L_COMMENT"`
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
	version := flag.Int("version", 0, "version to query for point query")
	start := flag.Int("start", 0, "start version for version query")
	end := flag.Int("end", 1, "end version for version query")
	flag.Parse()

	switch *transaction {
	case "BulkInvoke":
		BulkInvoke(contract, *file)
	case "BulkInvokeParallel":
		BulkInvokeParallel(contract, *file)
	case "Invoke":
		Invoke(contract, *file)
	case "getHistoryForAsset":
		getHistoryForAsset(contract, *key)
	case "getHistoryForAssets":
		getHistoryForAssets(contract, *key)
	case "getHistoryForAssetRange": // Add a new case for the new function
		getHistoryForAssetRange(contract, *key)
	case "getHistoryForAssetsOld":
		getHistoryForAssetsOld(contract, *key)
	case "pointQueryOld":
		pointQueryOld(contract, *key, *version)
	case "pointQuery":
		pointQuery(contract, *key, *version)
	case "versionQuery":
		versionQuery(contract, *key, *start, *end)
	}

}

func BulkInvoke(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		log.Fatalln("File URL must be provided and must be an absolute path")

	}

	jsonData, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	var table Table
	if err := json.Unmarshal([]byte(jsonData), &table); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	orders := table.Table

	startTime := time.Now()
	log.Printf("Starting bulk transaction at time: %s\n", startTime.Format(time.UnixDate))

	// Split orders into chunks of size 2500
	chunkSize := 2500
	for i := 0; i < len(orders); i += chunkSize {
		chunkTime := time.Now()

		chunk := orders[i:func() int {
			if i+chunkSize > len(orders) {
				return len(orders)
			}
			return i + chunkSize
		}()]

		chunkBytes, err := json.Marshal(chunk)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %s", err)
		}

		_, err = contract.SubmitTransaction("CreateBulk", string(chunkBytes))
		if err != nil {
			log.Fatalf("Failed to submit transaction: %s\n", err)
		}

		endTime := time.Now()
		executionTime := endTime.Sub(chunkTime).Seconds()
		log.Printf("Execution Time: %f sec at chunk %d", executionTime, i/chunkSize+1)
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

	raw, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		log.Fatalln(err)
	}

	var t Table
	json.Unmarshal(raw, &t)

	chunkSize := 500

	var wg sync.WaitGroup

	// Create a buffered channel to limit number of goroutines
	sem := make(chan bool, 10)

	for i := 0; i < len(t.Table); i += chunkSize {

		log.Printf("Processing chunk starting at index %d\n", i)

		end := i + chunkSize
		if end > len(t.Table) {
			end = len(t.Table)
		}
		chunk := t.Table[i:end]

		chunkBytes, err := json.Marshal(chunk)
		if err != nil {
			log.Println(err)
			continue
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

	jsonData, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		log.Fatalf("error while reading json file: %s", err)

	}

	var table Table
	if err := json.Unmarshal([]byte(jsonData), &table); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	orders := table.Table

	for i := 0; i < 10; i++ {

		orderBytes, err := json.Marshal(orders[i])
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %s", err)
		}

		_, err = contract.SubmitTransaction("Create", string(orderBytes))
		if err != nil {
			log.Fatalf("Failed to submit transaction: %s\n", err)
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

func getHistoryForAssetRange(contract *gateway.Contract, keys string) {
	startEndKeys := strings.Split(keys, ",")

	start, _ := strconv.Atoi(startEndKeys[0])
	end, _ := strconv.Atoi(startEndKeys[1])
	size := end - start + 1
	keys_list := make([]string, size)

	for i := range keys_list {
		keys_list[i] = strconv.Itoa(start + i)
	}

	result, err := contract.EvaluateTransaction("getHistoryForAssets", keys_list...)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	fmt.Println(string(result))
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

	if version < 0 || version >= len(assets) {
		log.Fatalf("Version number out of range: %d\n", version)
	}

	selectedAsset := assets[version]

	assetJSON, err := json.Marshal(selectedAsset)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s\n", err)
	}
	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	fmt.Println(string(assetJSON))
	log.Printf("Total execution time is: %f sec\n", executionTime)
}

func pointQuery(contract *gateway.Contract, key string, version int) {
	startTime := time.Now()

	versionString := strconv.Itoa(version)

	selectedAsset, err := contract.EvaluateTransaction("getVersionForAsset", key, versionString)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

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
func versionQuery(contract *gateway.Contract, key string, start int, end int) {
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

	if start < 0 || end < start || end >= len(assets) {
		log.Fatalf("Start or end index out of range: start=%d, end=%d\n", start, end)
	}

	selectedAssets := assets[start : end+1]

	assetsJSON, err := json.Marshal(selectedAssets)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %s\n", err)
	}
	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()
	fmt.Println(string(assetsJSON))
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
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
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
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
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
