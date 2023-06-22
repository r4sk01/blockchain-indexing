package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	TxId      string `json:"txId"`
	Value     string `json:"value"`
	Timestamp string `json:"timestamp"`
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

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Order
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) error {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "InitLedger":
		return s.InitLedger(stub)
	case "CreateBulk":
		return s.CreateBulk(stub, args)
	case "getHistoryForAsset":
		_, err := s.getHistoryForAsset(stub, args)
		return err
	default:
		return fmt.Errorf("invalid function name: %s. Expecting 'InitLedger', 'CreateBulk', or 'getHistoryForAsset'", function)
	}
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) error {
	fmt.Println("====== INIT CAllED ======")
	return nil
}

func (s *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) error {
	fmt.Println("====== INIT LEDGER START ======")
	fmt.Println("====== INIT LEDGER END ======")
	return nil
}

// CreateBulk Create a new key-value pair and send to state database
func (s *SmartContract) CreateBulk(stub shim.ChaincodeStubInterface, args []string) error {
	buffer := args[0]

	var orders []Order
	json.Unmarshal([]byte(buffer), &orders)

	for _, order := range orders {

		orderBytes, err := json.Marshal(order)
		if err != nil {
			return err
		}

		orderKey := strconv.FormatInt(int64(order.L_ORDERKEY), 10)

		log.Printf("Appending order %s with part %d\n", orderKey, order.L_PARTKEY)
		err = stub.PutState(orderKey, orderBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

// getHistoryForAsset executes getHistoryForKey() API
func (s *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	assetKey := args[0]
	historyResultsIterator, err := stub.GetHistoryForKey(assetKey)
	if err != nil {
		return "", fmt.Errorf("failed to get asset history for %s: %v", assetKey, err)
	}
	defer historyResultsIterator.Close()

	var history []Asset
	for historyResultsIterator.HasNext() {
		response, err := historyResultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to iterate asset history: %v", err)
		}

		txTimestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return "", fmt.Errorf("failed to parse timestamp: %v", err)
		}

		asset := Asset{
			TxId:      response.TxId,
			Value:     string(response.Value),
			Timestamp: txTimestamp.String(),
		}

		history = append(history, asset)
	}

	historyAsJSON, err := json.Marshal(history)
	if err != nil {
		return "", fmt.Errorf("failed to marshal asset history: %v", err)
	}
	return string(historyAsJSON), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
