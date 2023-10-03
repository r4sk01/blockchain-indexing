#!/bin/bash

# go run application.go -t versionQueryOld -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 199

# go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/blockTransactions17000000-17010000.json

files=("blockTransactions17125001-17150000.json")

dataDir="/home/andrey/Documents/insert-tpch/ethereumData/Second100K/"

for file in "${files[@]}"; do
    fullPath="$dataDir$file"
    if [[ -f "$fullPath" ]]; then
        echo "Inserting file: $fullPath"
        go run application.go -t BulkInvokeParallel -f "$fullPath"
    else
        echo "Warning: $fullPath not found."
    fi
done

echo "INSERTION SECOND 100K 2-2 DONE"