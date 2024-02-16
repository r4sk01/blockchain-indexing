#!/bin/bash

results=insertResults-original.txt

echo "Inserting 1 Million" >> "$results"

for ((i = 0; i < 3; i++)); do
    echo ./original-startFabric.sh go
    pushd go
    go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17000000-17010000.json >> "$results" 2>&1
    popd
    echo ./original-networkDown.sh
done
