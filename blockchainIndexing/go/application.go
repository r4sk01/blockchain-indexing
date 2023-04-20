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
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
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
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	contract := network.GetContract("blockchainIndexing")

	transaction := flag.String("t", "defaultQuery", "Choose a transaction to run")
	file := flag.String("f", "~", "file path for json data")
	flag.Parse()

	switch *transaction {
	case "bulkInvoke":
		bulkInvoke(contract, *file)
	/*case "stateRangeQuery":
	stateRangeQuery(contract)*/
	case "histTest":
		histTest(contract)
	case "pointQuery":
		pointQuery(contract)
	case "versionQuery":
		versionQuery(contract)
	case "rangeQuery":
		rangeQuery(contract)
	}

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

func bulkInvoke(contract *gateway.Contract, fileUrl string) {
	if fileUrl == "" || !filepath.IsAbs(fileUrl) {
		fmt.Println("File URL must be provided and must be an absolute path")
		os.Exit(1)
	}

	// Read JSON file
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

	// Split orders into chunks of size 1000
	chunkSize := 10000
	for index := 0; index < len(orders); index += chunkSize {
		end := index + chunkSize
		if end > len(orders) {
			end = len(orders)
		}
		chunk := orders[index:end]
		if index == 20000 {
			break
		}

		// Marshal chunk into JSON string
		jsonString, err := json.Marshal(chunk)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %s", err)
		}

		startTime := time.Now()
		_, err = contract.SubmitTransaction("CreateBulk", string(jsonString))
		if err != nil {
			log.Fatalf("Failed to submit transaction: %s\n", err)

		}
		endTime := time.Now()
		executionTime := endTime.Sub(startTime).Milliseconds()
		log.Printf("Execution Time: %d ms at chunk %d", executionTime, index+10000)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Milliseconds()
	log.Printf("Finished bulk transaction at time: %s\n", endTime.Format(time.UnixDate))
	log.Printf("Total execution time is: %d ms\n", executionTime)

}

/*
func stateRangeQuery(contract *gateway.Contract) {
	log.Println("-----Range Query Orders-----")
	startKey := "91041"
	endKey := "970757"

	result, err := contract.EvaluateTransaction("ReadByKeyRange", startKey, endKey)
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transaction has been evaluated, result is: %s\n", string(result))

}*/

func histTest(contract *gateway.Contract) {
	log.Println("-----Query Order-----")

	key := "24454"

	result, err := contract.EvaluateTransaction("histTest", key)
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transaction has been evaluated, result is: %s\n", string(result))
}

func pointQuery(contract *gateway.Contract) {
	startTime := time.Now()

	key := "9000"
	version := "3"

	_, err := contract.EvaluateTransaction("pointQuery", key, version)
	if err != nil {
		fmt.Println("Failed to evaluate transaction")
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Milliseconds()
	log.Printf("Finished point query with execution time: %d ms\n", executionTime)

}

func versionQuery(contract *gateway.Contract) {
	startTime := time.Now()

	key := "9000"
	startVer := "5"
	endVer := "15"

	_, err := contract.EvaluateTransaction("versionQuery", key, startVer, endVer)
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s ms\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Milliseconds()
	log.Printf("Finished point query with execution time: %d ms\n", executionTime)

}

func rangeQuery(contract *gateway.Contract) {
	startTime := time.Now()

	startKey := "1000"
	endKey := "10000"

	_, err := contract.EvaluateTransaction("rangeQuery", startKey, endKey)
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
	}

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Milliseconds()
	log.Printf("Finished point query with execution time: %d ms\n", executionTime)

}
