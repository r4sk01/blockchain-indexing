package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
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
	case "getHistoryForAsset":
		return sc.getHistoryForAsset(stub, args)
	// Requires GetHistoryForKeys API
	case "getHistoryForAssets":
		return sc.getHistoryForAssets(stub, args)
	// Requires GetVersionsForKey API
	case "getVersionsForAsset":
		return sc.getVersionsForAsset(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
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

	keyBytes, err := hex.DecodeString(transaction.From[2:])
	if err != nil {
		return shim.Error("Error: could not decode hex from string: " + transaction.From)
	}
	transactionKey := string(keyBytes)
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

		keyBytes, err := hex.DecodeString(transaction.From[2:])
		if err != nil {
			return shim.Error("Error: could not decode hex from string: " + transaction.From)
		}
		transactionKey := string(keyBytes)

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

		keyBytes, err := hex.DecodeString(transaction.From[2:])
		if err != nil {
			return shim.Error("Error: could not decode hex from string: " + transaction.From)
		}
		transactionKey := string(keyBytes)
		err = stub.PutState(transactionKey, transactionBytes)
		if err != nil {
			return shim.Error("Failed to create transaction: " + err.Error())
		}
	}
	return shim.Success(nil)
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

// getHistoryForAssets calls custom GetHistoryForKeys() API
func (sc *SmartContract) getHistoryForAssets(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 or more")
	}

	// calling the GetHistoryForKeys() API with keys as args
	historyItr, err := stub.GetHistoryForKeys(args) // historyIters in old version
	if err != nil {
		return shim.Error(err.Error())
	}

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

	// var histories [][]QueryResult
	// for _, historyItr := range historyItrs {
	// 	var history []QueryResult
	// 	for historyItr.HasNext() {
	// 		historyData, err := historyItr.Next()
	// 		if err != nil {
	// 			return shim.Error(err.Error())
	// 		}

	// 		var transaction Transaction
	// 		json.Unmarshal(historyData.Value, &transaction)

	// 		history = append(history, QueryResult{Key: historyData.TxId, Record: &transaction})
	// 	}
	// 	histories = append(histories, history)
	// }

	// historiesAsBytes, _ := json.Marshal(histories)
	// return shim.Success(historiesAsBytes)

	historyAsBytes, _ := json.Marshal(history)
	return shim.Success(historyAsBytes)
}

func (sc *SmartContract) getVersionsForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	start, _ := strconv.ParseUint(args[1], 10, 64)
	end, _ := strconv.ParseUint(args[2], 10, 64)

	versionIter, err := stub.GetVersionsForKey(args[0], start, end)
	if err != nil {
		return shim.Error(err.Error())
	}

	var versions []QueryResult
	for versionIter.HasNext() {
		versionData, err := versionIter.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var transaction Transaction
		json.Unmarshal(versionData.Value, &transaction) // .Value?

		timestamp := time.Unix(versionData.Timestamp.Seconds, int64(versionData.Timestamp.Nanos)).String()

		versions = append(versions, QueryResult{Key: versionData.TxId, Record: &transaction, Timestamp: timestamp})
	}

	versionAsBytes, _ := json.Marshal(versions)
	return shim.Success(versionAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
