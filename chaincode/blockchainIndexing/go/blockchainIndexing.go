package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/RUAN0007/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
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

type QueryResult struct {
	Key       string       `json:"Key"`
	Record    *Transaction `json:"record"`
	Timestamp string       `json:"timestamp"`
}

// SimpleContract contract for handling writing and reading from the world state
type SmartContract struct {
}

func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	switch function {
	case "InitLedger":
		return sc.InitLedger(stub)
	case "CreateBulk":
		return sc.CreateBulk(stub, args)
	case "CreateBulkParallel":
		return sc.CreateBulkParallel(stub, args)
	case "Create":
		return sc.Create(stub, args)
	case "PointQuery":
		return sc.PointQuery(stub, args)
	case "VersionQuery":
		return sc.VersionQuery(stub, args)
	case "RangeQuery":
		return sc.RangeQuery(stub, args)
	case "HistTest":
		return sc.HistTest(stub, args)
	case "getHistoryForAsset":
		return sc.getHistoryForAsset(stub, args)
	case "getState":
		return sc.getState(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (sc *SmartContract) Prov(stub shim.ChaincodeStubInterface, reads, writes map[string][]byte) map[string][]string {
	var dependecies map[string][]string

	return dependecies
}

func (sc *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) sc.Response {
	log.Println("'============= Initialized Ledger ==========='")
	return shim.Success(nil)

}

func (sc *SmartContract) Create(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var transaction Transaction
	json.Unmarshal([]byte(args[0]), &transaction)

	transactionBytes, err := json.Marshal(transaction)
	if err != nil {
		return shim.Error("Failed to marshal transaction JSON: " + err.Error())
	}

	transactionKey := transaction.From
	log.Printf("Appending transaction: %s\n", transactionKey)

	err = stub.PutState(transactionKey, transactionBytes)
	if err != nil {
		return shim.Error("failed to put transaction on ledger: " + err.Error())
	}

	return shim.Success(nil)

}

// Create a new key-value pair and send to state database
func (sc *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var transactions []Transaction
	json.Unmarshal([]byte(buffer), &transactions)

	for _, transaction := range transactions {

		transactionBytes, err := json.Marshal(transaction)
		if err != nil {
			return shim.Error("failed to marshal transaction JSON: " + err.Error())
		}

		transactionKey := transaction.From

		// Fabric key must be a string
		//fmt.Sprintf("%d", transaction.L_ORDERKEY)
		log.Printf("Appending transaction %s with gasPrice %d\n", transactionKey, transaction.GasPrice)
		err = stub.PutState(transactionKey, transactionBytes)
		if err != nil {
			return shim.Error("failed to put transaction on ledger: " + err.Error())
		}
	}

	return shim.Success(nil)

}

func (sc *SmartContract) CreateBulkParallel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var transactions []Transaction
	json.Unmarshal([]byte(args[0]), &transactions)

	for _, transaction := range transactions {
		transactionBytes, err := json.Marshal(transaction)
		if err != nil {
			return shim.Error("Error marshaling transaction object: " + err.Error())
		}

		err = stub.PutState(transaction.From, transactionBytes)
		if err != nil {
			return shim.Error("Failed to create transaction: " + err.Error())
		}
	}
	return shim.Success(nil)
}

func (sc *SmartContract) PointQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]
	version, _ := strconv.ParseUint(args[1], 10, 64)
	startBlk, _ := strconv.ParseUint(args[2], 10, 64)
	endBlk, _ := strconv.ParseUint(args[3], 10, 64)
	var results []string

	for startBlk <= endBlk {
		val, _, err := stub.Hist(key, startBlk)
		if err != nil {
			shim.Error("Failed to get historical value: " + err.Error())
		}

		results = append(results, val)
		startBlk++

	}

	log.Printf("pointQuery results: %s\n", results)

	resultsBytes, err := json.Marshal(results[version])
	if err != nil {
		shim.Error("Marhsal failed with: " + err.Error())
	}

	return shim.Success(resultsBytes)

}

func (sc *SmartContract) VersionQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]
	startVersion, _ := strconv.ParseUint(args[1], 10, 64)
	endVersion, _ := strconv.ParseUint(args[2], 10, 64)
	startBlk, _ := strconv.ParseUint(args[3], 10, 64)
	endBlk, _ := strconv.ParseUint(args[4], 10, 64)
	var results []string

	for startBlk <= endBlk {
		val, _, err := stub.Hist(key, startBlk)
		if err != nil {
			shim.Error("Failed to get historical value: " + err.Error())
		}

		results = append(results, val)
		startBlk++

	}

	log.Printf("versionQuery results: %s\n", results)

	resultsBytes, err := json.Marshal(results[startVersion : endVersion+1])
	if err != nil {
		shim.Error("Marhsal failed with: " + err.Error())
	}

	return shim.Success(resultsBytes)
}

func (sc *SmartContract) RangeQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	startKey, _ := strconv.ParseUint(args[0], 10, 64)
	endKey, _ := strconv.ParseUint(args[1], 10, 64)
	startBlk, _ := strconv.ParseUint(args[2], 10, 64)
	endBlk, _ := strconv.ParseUint(args[3], 10, 64)
	var results []string

	for key := startKey; key <= endKey; key++ {
		keyStr := strconv.FormatUint(key, 10)
		log.Println(keyStr)
		for startBlk <= endBlk {
			val, _, err := stub.Hist(keyStr, startBlk)
			if err != nil {
				shim.Error("Failed to get historical value: " + err.Error())
			}

			results = append(results, val)
			startBlk++

		}
		startBlk, _ = strconv.ParseUint(args[0], 10, 64)

	}

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}
	return shim.Success(resultsBytes)

}

func (sc *SmartContract) HistTest(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	log.Println("-----Hist Test-----")
	startKey, _ := strconv.ParseUint(args[0], 10, 64)
	endKey, _ := strconv.ParseUint(args[1], 10, 64)
	startBlk, _ := strconv.ParseUint(args[2], 10, 64)
	endBlk, _ := strconv.ParseUint(args[3], 10, 64)
	var results []string

	for key := startKey; key <= endKey; key++ {
		keyStr := strconv.FormatUint(key, 10)
		log.Println(keyStr)
		for startBlk <= endBlk {
			val, _, err := stub.Hist(keyStr, startBlk)
			if err != nil {
				shim.Error("Failed to get historical value: " + err.Error())
			}

			results = append(results, val)
			startBlk++

		}
		startBlk, _ = strconv.ParseUint(args[0], 10, 64)

	}

	log.Println(results)

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}
	return shim.Success(resultsBytes)

}

// getHistoryForAsset calls built in GetHistoryForKey() API
func (sc *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	historyItr, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyItr.Close()

	var history []QueryResult
	for historyItr.HasNext() {
		historyData, err := historyItr.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var transaction Transaction
		json.Unmarshal(historyData.Value, &transaction)

		//Convert google.protobuf.Timestamp to string
		timestamp := time.Unix(historyData.Timestamp.Seconds, int64(historyData.Timestamp.Nanos)).String()

		history = append(history, QueryResult{Key: historyData.TxId, Record: &transaction, Timestamp: timestamp})
	}

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

func (sc *SmartContract) getState(stub shim.ChaincodeStubInterface, args []string) Transaction {
	log.Println("-----Hist Test-----")
	key := args[0]

	val, _ := stub.GetState(key)

	var tx Transaction
	json.Unmarshal(val, &tx)
	return tx

	// log.Println(val)

	// resultsBytes, err := json.Marshal(val)
	// if err != nil {
	// 	return shim.Error("failed to marshal order JSON: " + err.Error())
	// }
	// return shim.Success(resultsBytes)

}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
