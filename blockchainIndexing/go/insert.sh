#!/bin/bash

# go run application.go -t versionQueryOld -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 199

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17000000-17010000.json

# go run application.go -t getHistoryForAssetPaginated -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -p 10

files=("blockTransactions17000000-17010000.json" "blockTransactions17010001-17011000.json" "blockTransactions17011001-17012000.json" \ 
    "blockTransactions17012001-17015000.json" "blockTransactions17015001-17020000.json" "blockTransactions17020001-17030000.json" \
    "firstHalf.json" "secondHalf.json" \
    "blockTransactions17090001-17100000.json")

dataDir="/home/andrey/Documents/insert-tpch/ethereum"

for file in ${files[@]}; do
    echo "Inserting file: $dataDir/$file"
    go run application.go -t BulkInvokeParallel -f "$dataDir/$file" 
done