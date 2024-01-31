#!/bin/bash

results=mapResults.txt

> "$results"

echo "Inserting 1 Million" >> "$results"
go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17000000-17010000.json

for ((i = 0; i < 6; i++)); do
    go run application.go -t blockRangeQuery -start 1000 -end 1200 -u 500 >> "$results"
done

for ((i = 0; i < 6; i++)); do
    go run application.go -t blockRangeQuery -start 1000 -end 1500 -u 1000 >> "$results"
done

echo "" >> "$results"
echo "Inserting 10 Million: " >> "$results"

filenames=(
"First100K/blockTransactions17010001-17011000.json" 
"First100K/blockTransactions17011001-17012000.json" 
"First100K/blockTransactions17012001-17015000.json"
"First100K/blockTransactions17015001-17020000.json"
"First100K/blockTransactions17020001-17030000.json"
"First100K/blockTransactions17030001-17050000.json"
"First100K/blockTransactions17090001-17100000.json"
)

dataDir="/home/andrey/Documents/insert-tpch/ethereum"

for file in ${filenames[@]}; do
    echo "Inserting file: $dataDir/$file"
    go run application.go -t BulkInvokeParallel -f "$dataDir/$file" 
done

for ((i = 0; i < 6; i++)); do
    go run application.go -t blockRangeQuery -start 1000 -end 1200 -u 500 >> "$results"
done

for ((i = 0; i < 6; i++)); do
    go run application.go -t blockRangeQuery -start 1000 -end 1500 -u 1000 >> "$results"
done
