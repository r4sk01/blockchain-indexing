package main

import (
	"encoding/json"
	"fmt"
	"math"
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

type HistResult struct {
	Msg        string
	Val        string
	CreatedBlk uint64
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
	/*case "ReadByKeyRange":
	return sc.ReadByKeyRange(stub, args)*/
	case "pointQuery":
		return sc.pointQuery(stub, args)
	case "versionQuery":
		return sc.versionQuery(stub, args)
	case "rangeQuery":
		return sc.rangeQuery(stub, args)
	case "histTest":
		return sc.histTest(stub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

}

func (sc *SmartContract) Prov(stub shim.ChaincodeStubInterface, reads, writes map[string][]byte) map[string][]string {
	return nil
}

func (sc *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("'============= Initialized Ledger ==========='")
	return shim.Success(nil)

}

// Create adds a new key with value to the world state
func (sc *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	buffer := args[0]

	var orders []Order
	json.Unmarshal([]byte(buffer), &orders)

	for _, order := range orders {

		orderBytes, err := json.Marshal(order)
		if err != nil {
			return shim.Error("failed to marshal order JSON: " + err.Error())
		}

		// Fabric key must be a string
		fmt.Printf("Appending order: %d\n", order.L_ORDERKEY)
		err = stub.PutState(fmt.Sprintf("%d", order.L_ORDERKEY), orderBytes)
		if err != nil {
			return shim.Error("failed to put order on ledger: " + err.Error())
		}
	}

	return shim.Success(nil)

}

func (sc *SmartContract) histTest(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]
	var blk uint64 = math.MaxUint64

	val, comBlk, err := stub.Hist(key, blk)
	if err != nil {
		return shim.Error("Hist() failed with: " + err.Error())
	}

	fmt.Printf("Val %s at block height %d\n", val, comBlk)

	valBytes, err := json.Marshal(val)
	if err != nil {
		return shim.Error("Marshal failed with: " + err.Error())
	}

	return shim.Success(valBytes)

}

/* Read returns the value at key in the world state
func (sc *SmartContract) ReadByKeyRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	startKey := args[0]
	endKey := args[1]
	results := []QueryResult{}

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		order := new(Order)
		_ = json.Unmarshal(queryResponse.Value, order)
		fmt.Printf("Order: %+v\n", order)

		queryResult := QueryResult{Key: queryResponse.Key, Record: order}
		results = append(results, queryResult)
	}

	resultsAsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}

	return shim.Success(resultsAsBytes)
}*/

func (sc *SmartContract) pointQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]
	blk, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		shim.Error(err.Error())
	}

	val, _, err := stub.Hist(key, blk)
	if err != nil {
		return shim.Error("Failed to get historical value: " + err.Error())
	}

	return shim.Success([]byte(val))

}

func (sc *SmartContract) versionQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var results []string
	key := args[0]
	blkStart, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	blkEnd, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}

	for b := blkStart; b <= blkEnd; b++ {
		record, _, err := stub.Hist(key, b)
		if err != nil {
			continue
		}
		results = append(results, record)
	}

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}

	return shim.Success(resultsBytes)

}

func (sc *SmartContract) rangeQuery(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	startKey, _ := strconv.ParseUint(args[0], 10, 64)
	endKey, _ := strconv.ParseUint(args[1], 10, 64)
	results := []Order{}

	for k := startKey; k <= endKey; k++ {
		key := strconv.FormatUint(k, 10)
		iterator, err := stub.GetHistoryForKey(key)
		if err != nil {
			fmt.Printf("Historical query failed: %s\n", err)
			continue
		}

		defer iterator.Close()
		for iterator.HasNext() {
			response, err := iterator.Next()
			if err != nil {
				return shim.Error(err.Error())
			}

			var order Order
			_ = json.Unmarshal(response.Value, &order)
			results = append(results, order)

		}
	}

	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error("failed to marshal order JSON: " + err.Error())
	}
	return shim.Success(resultsBytes)

}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting chaincode: %v \n", err)
	}
}
