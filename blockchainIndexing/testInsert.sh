#!/bin/bash

results=insertResults-TPCH-1M.txt


echo "" >> "$results"
echo "ORIGINAL" >> "$results"

echo "" >> "$results"
echo "PARALLEL" >> "$results"
for ((i = 0; i < 3; i++)); do
    ./startFabric.sh go
    sleep 10
    pushd go

    for file in ${filenames[@]}; do
        echo "Inserting file: /home/andrey/Documents/insert-tpch/sortUnsort10500/unsortedMilEntries.json"
        go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/sortUnsort10500/unsortedMilEntries.json >> ../"$results" 2>&1
    done

    popd
    ./networkDown.sh
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