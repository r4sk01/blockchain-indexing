package main

import (
	"encoding/json"
	"github.com/RUAN0007/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"log"
	"strconv"
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
	Key       string `json:"Key"`
	Record    *Order `json:"record"`
	Timestamp string `json:"timestamp"`
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

func (sc *SmartContract) CreateBulkParallel(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var orders []Order
	json.Unmarshal([]byte(args[0]), &orders)

	for _, order := range orders {
		orderBytes, err := json.Marshal(order)
		if err != nil {
			return shim.Error("Error marshaling order object: " + err.Error())
		}

		err = stub.PutState(strconv.Itoa(order.L_ORDERKEY), orderBytes)
		if err != nil {
			return shim.Error("Failed to create order: " + err.Error())
		}
	}
	return shim.Success(nil)
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

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Printf("Error starting chaincode: %v \n", err)
	}
}
