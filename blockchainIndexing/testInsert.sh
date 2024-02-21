#!/bin/bash

results=insertResults-10M.txt

filenames=(
"blockTransactions17000000-17010000.json"
"blockTransactions17010001-17011000.json" 
"blockTransactions17011001-17012000.json" 
"blockTransactions17012001-17015000.json"
"blockTransactions17015001-17020000.json"
"blockTransactions17020001-17030000.json"
"blockTransactions17030001-17050000.json"
"blockTransactions17090001-17100000.json"
)

dataDir="/home/andrey/Documents/insert-tpch/ethereum/First100K"

echo "" >> "$results"
echo "BLOCK" >> "$results"

echo "" >> "$results"
echo "PARALLEL" >> "$results"
for ((i = 0; i < 3; i++)); do
    ./original-startFabric.sh go
    sleep 10
    pushd go

    echo "" >> ../"$results"
    for file in ${filenames[@]}; do
        echo "Inserting file: $dataDir/$file"
        go run application.go -t BulkInvokeParallel -f "$dataDir/$file" >> ../"$results" 2>&1
    done

    popd
    ./original-networkDown.sh
done

# echo "" >> "$results"
# echo "SEQUENTIAL" >> "$results"
# ./original-startFabric.sh go
# sleep 10
# pushd go

# for file in ${filenames[@]}; do
#     echo "Inserting file: $dataDir/$file"
#     go run application.go -t BulkInvoke -f "$dataDir/$file" >> ../"$results" 2>&1
# done

# popd
# ./original-networkDown.sh