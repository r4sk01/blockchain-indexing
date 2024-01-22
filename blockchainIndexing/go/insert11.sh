#!/bin/bash

# go run application.go -t versionQuery -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17010001-17011000.json

# go run application.go -t getHistoryForAsset -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 
# "blockTransactions17000000-17010000.json"
# docker exec -it peer0.org1.example.com sh

files1=("blockTransactions17000000-17010000.json"
        "blockTransactions17010001-17011000.json"
        "blockTransactions17011001-17012000.json"
        "blockTransactions17012001-17015000.json")

dataDir1="/home/andrey/Documents/insert-tpch/ethereumData/First100K"

for file in ${files1[@]}; do
    echo "Inserting file: $file"
    go run application.go -t BulkInvokeParallel -f "$dataDir1/$file" 
done