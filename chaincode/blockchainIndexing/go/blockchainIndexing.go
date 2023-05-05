package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/RUAN0007/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

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

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Order
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
	case "Create":
		return sc.Create(stub, args)
	case "pointQuery":
		return sc.pointQuery(stub, args)
	case "versionQuery":
		return sc.versionQuery(stub, args)
	case "rangeQuery":
		return sc.rangeQuery(stub, args)
	case "histTest":
		return sc.histTest(stub, args)
	case "CreateBulkPL":
		return sc.CreateBulkPL(stub, args)
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
	var order Order
	json.Unmarshal([]byte(args[0]), &order)

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return shim.Error("Failed to marshal order JSON: " + err.Error())
	}

	orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)
	log.Printf("Appending order: %s\n", orderKey)

	err = stub.PutState(orderKey, orderBytes)
	if err != nil {
		return shim.Error("failed to put order on ledger: " + err.Error())
	}

	return shim.Success(nil)

}

// Create a new key-value pair and send to state database
func (sc *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var orders []Order
	json.Unmarshal([]byte(buffer), &orders)

	for _, order := range orders {

		orderBytes, err := json.Marshal(order)
		if err != nil {
			return shim.Error("failed to marshal order JSON: " + err.Error())
		}

		orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)

		// Fabric key must be a string
		//fmt.Sprintf("%d", order.L_ORDERKEY)
		log.Printf("Appending order %s with part %d\n", orderKey, order.L_PARTKEY)
		err = stub.PutState(orderKey, orderBytes)
		if err != nil {
			return shim.Error("failed to put order on ledger: " + err.Error())
		}
	}

	return shim.Success(nil)

}

// Create a new key-value pair and send to state database
func (sc *SmartContract) CreateBulkPL(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var orders []Order
	json.Unmarshal([]byte(buffer), &orders)

	// Create a channel to receive errors from Goroutines.
	errChan := make(chan error, len(orders))

	// Process each order in a separate Goroutine.
	for _, order := range orders {
		go func(order Order) {
			orderBytes, err := json.Marshal(order)
			if err != nil {
				errChan <- fmt.Errorf("failed to marshal order JSON: %v", err)
				return
			}

			orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)

			log.Printf("Appending order %s with part %d\n", orderKey, order.L_PARTKEY)

			err = stub.PutState(orderKey, orderBytes)
			if err != nil {
				errChan <- fmt.Errorf("failed to put order on ledger: %v", err)
				return
			}

			errChan <- nil
		}(order)
	}

	// Wait for all Goroutines to finish.
	for i := 0; i < len(orders); i++ {
		if err := <-errChan; err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)

}

// test function for stub.Hist()
func (sc *SmartContract) histTest(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]
	startBlk, _ := strconv.ParseUint(args[1], 10, 64)
	endBlk, _ := strconv.ParseUint(args[2], 10, 64)
	var results []string

	for startBlk <= endBlk {
		val, _, err := stub.Hist(key, startBlk)
		if err != nil {
			shim.Error("Failed to get historical value: " + err.Error())
		}

		results = append(results, val)
		startBlk++

	}

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		shim.Error("Marhsal failed with: " + err.Error())
	}

	return shim.Success(resultsBytes)

}

// Obtain a single version for a single key
func (sc *SmartContract) pointQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
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

// Obtain a range of versions for a single key
func (sc *SmartContract) versionQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
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

// Obtain all versions for a range of keys
func (sc *SmartContract) rangeQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	startKey := args[0]
	endKey := args[1]
	startBlk, _ := strconv.ParseUint(args[2], 10, 64)
	endBlk, _ := strconv.ParseUint(args[3], 10, 64)
	var results []string

	iterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error("Error getting key range: " + err.Error())
	}

	defer iterator.Close()

	for iterator.HasNext() {

		for startBlk <= endBlk {
			currKey, err := iterator.Next()
			if err != nil {
				return shim.Error("Error getting next key from iterator: " + err.Error())
			}

			val, _, err := stub.Hist(currKey.Key, startBlk)
			if err != nil {
				shim.Error("Failed to get historical value: " + err.Error())
			}

			results = append(results, val)
			startBlk++

		}

	}

	log.Printf("rangeQuery results: %s\n", results)

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}
	return shim.Success(resultsBytes)

}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
