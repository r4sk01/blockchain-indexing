#!/bin/bash

results=insertResults-TPCH-1M.txt

dataFile=/home/andrey/Documents/insert-tpch/sortUnsort10500/unsortedMilEntries.json

echo "" >> "$results"
echo "BLOCK" >> "$results"

echo "" >> "$results"
echo "PARALLEL" >> "$results"
for ((i = 0; i < 3; i++)); do
    ./startFabric.sh go
    sleep 10
    pushd go

    echo "Inserting file: $dataFile"
    go run application.go -t BulkInvokeParallel -f "$dataFile" >> ../"$results" 2>&1

    if (( i == 2 )); then
        echo "" >> ../"$results"
        for ((j=0; j < 6; j++)); do
            go run application.go -t blockRangeQuery -start 100 -end 360 -u 7 >> ../"$results" 2>&1
        done
    fi

    popd
    ./networkDown.sh
done

# echo "" >> "$results"
# echo "SEQUENTIAL" >> "$results"
# ./startFabric.sh go
# sleep 10
# pushd go

# echo "Inserting file: $dataFile"
# go run application.go -t BulkInvoke -f "$dataFile" >> ../"$results" 2>&1

# popd
# ./networkDown.sh