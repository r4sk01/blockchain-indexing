package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

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
	flag.Parse()

	switch *transaction {
	case "BulkInvoke":
		BulkInvoke(contract, *file)
	case "Invoke":
		Invoke(contract, *file)
	case "getHistoryForAsset": // Add a new case for the new function
		getHistoryForAsset(contract, *key)
		// case "getHistoryForAssets": // Add a new case for the new function
		// 	getHistoryForAssets(contract, *key)
	}

}

func BulkInvoke(contract *gateway.Contract, fileUrl string) {
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

	// Create a wait group to synchronize goroutines
	var wg sync.WaitGroup

	for i := 0; i < len(t.Table); i += chunkSize {
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
		go func(data string) {
			defer wg.Done()
			_, err = contract.SubmitTransaction("CreateBulk", data)
			if err != nil {
				log.Println(err)
			}
		}(string(chunkBytes))
	}

	// Wait for all goroutines to finish
	wg.Wait()
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
	result, err := contract.EvaluateTransaction("getHistoryForAsset", key)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %s\n", err)
	}

	fmt.Println(string(result))
}

// func getHistoryForAssets(contract *gateway.Contract, key string) {
// 	result, err := contract.EvaluateTransaction("getHistoryForAssets", key)
// 	if err != nil {
// 		log.Fatalf("Failed to evaluate transaction: %s\n", err)
// 	}

// 	fmt.Println(string(result))
// }

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
