#!/bin/bash

 

# Arrays of files
files1=(
    "blockTransactions17000000-17010000.json"
    "blockTransactions17010001-17011000.json"
    "blockTransactions17011001-17012000.json"
    "blockTransactions17012001-17015000.json"
    "blockTransactions17015001-17020000.json"
    "blockTransactions17020001-17030000.json"
    "blockTransactions17030001-17050000.json"
    "blockTransactions17090001-17100000.json"
)

 

files2=(
    "blockTransactions17100000-17125000.json"
    "blockTransactions17125001-17150000.json"
    "blockTransactions17150001-17175000.json"
)

 

files3=(
    "blockTransactions17200000-17225000.json"
    "blockTransactions17225001-17250000.json"
    "blockTransactions17250001-17275000.json"
    "blockTransactions17275001-17300000.json"
)

 

# Directories
dataDir1="/home/andrey/Documents/insert-tpch/ethereumData/First100K"
dataDir2="/home/andrey/Documents/insert-tpch/ethereumData/Second100K"
dataDir3="/home/andrey/Documents/insert-tpch/ethereumData/Third100K"

 

# Function to process files
process_files() {
    local dataDir="$1"
    shift
    local files=("$@")

 

    for file in "${files[@]}"; do
        echo "Inserting file: $dataDir/$file"
        go run application.go -t BulkInvokeParallel -f "$dataDir/$file"
        if [ $? -ne 0 ]; then
            echo "Error processing file: $dataDir/$file"
            exit 1
        fi
    done
}

 

# Process the files
process_files "$dataDir1" "${files1[@]}"
process_files "$dataDir2" "${files2[@]}"
process_files "$dataDir3" "${files3[@]}"