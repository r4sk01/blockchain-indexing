#!/bin/bash

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/First100K/

filenames=(
#"First100K/blockTransactions17000000-17010000.json" 
"First100K/blockTransactions17010001-17011000.json" 
"First100K/blockTransactions17011001-17012000.json" 
"First100K/blockTransactions17012001-17015000.json"
"First100K/blockTransactions17015001-17020000.json"
"First100K/blockTransactions17020001-17030000.json"
"First100K/blockTransactions17030001-17050000.json"
"First100K/blockTransactions17090001-17100000.json"
# "Second100K/blockTransactions17100000-17125000.json"
# "Second100K/blockTransactions17125001-17150000.json"
# "Second100K/blockTransactions17150001-17175000.json"
# "Third100K/blockTransactions17200000-17225000.json"
# "Third100K/blockTransactions17225001-17250000.json"
# "Third100K/blockTransactions17250001-17275000.json"
# "Third100K/blockTransactions17275001-17300000.json"
)

dataDir="/home/andrey/Documents/insert-tpch/ethereum"

for file in ${filenames[@]}; do
    echo "Inserting file: ${dataDir}/${file}"
    go run application.go -t BulkInvokeParallel -f "${dataDir}/${file}" 
done