#!/bin/bash

# go run application.go -t versionQueryOld -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 199

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17010001-17011000.json

# go run application.go -t getHistoryForAssetPaginated -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -p 10

files1=("blockTransactions17000000-17010000.json" "blockTransactions17010001-17011000.json" "blockTransactions17011001-17012000.json" \ 
    "blockTransactions17012001-17015000.json" "blockTransactions17015001-17020000.json" "blockTransactions17020001-17030000.json" \
    "firstHalf.json" "secondHalf.json" \
    "blockTransactions17090001-17100000.json")

# Broken: blockTransactions17175001-17200000.json
files2=("blockTransactions17125001-17150000.json" "blockTransactions17100000-17125000.json" \
        "blockTransactions17150001-17175000.json" )

files3=("blockTransactions17200000-17225000.json" "blockTransactions17225001-17250000.json" \
        "blockTransactions17250001-17275000.json" "blockTransactions17275001-17300000.json")

dataDir1="/home/andrey/Documents/insert-tpch/ethereum"
dataDir2="/home/andrey/Documents/insert-tpch/ethereum/Second100K"
dataDir3="/home/andrey/Documents/insert-tpch/ethereum/Third100K"

for file in ${files2[@]}; do
    echo "Inserting file: $dataDir2/$file"
    go run application.go -t BulkInvokeParallel -f "$dataDir2/$file" 
done

for file in ${files3[@]}; do
    echo "Inserting file: $dataDir3/$file"
    go run application.go -t BulkInvokeParallel -f "$dataDir3/$file" 
done

for file in ${files1[@]}; do
    echo "Inserting file: $dataDir1/$file"
    go run application.go -t BulkInvokeParallel -f "$dataDir1/$file" 
done