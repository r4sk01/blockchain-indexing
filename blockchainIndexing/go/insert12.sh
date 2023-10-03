#!/bin/bash

# go run application.go -t versionQueryOld -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 199

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17000000-17010000.json

files=("blockTransactions17015001-17020000.json" "blockTransactions17020001-17030000.json")

dataDir="/home/andrey/Documents/insert-tpch/ethereumData/First100K/"

#for file in ${files[@]}; do
for file in $dataDir*; do
    echo "Inserting file: $file"
    go run application.go -t BulkInvokeParallel -f "$file" 
done

echo "INSERTION FIRST 100K 1-2 DONE"