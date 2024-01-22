#!/bin/bash

# go run application.go -t versionQuery -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17010001-17011000.json

# go run application.go -t getHistoryForAsset -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 

files1=("blockTransactions17275001-17300000.json")

dataDir1="/home/andrey/Documents/insert-tpch/ethereumData/Third100K"

for file in ${files1[@]}; do
    echo "Inserting file: $file"
    go run application.go -t BulkInvokeParallel -f "$dataDir1/$file" 
done