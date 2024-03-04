#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version, insert data, run range query tests
#
# Author: Daniel Garon
# Date: 2024-02-21 
# Checked with shellcheck.net

main() {
    local results=/home/andrey/Desktop/key_range_test.txt
    local branches=(
        dgaron-2.3-blockRangeQuery-OriginalIndex
        dgaron-2.3-blockRangeQuery-VBI
        dgaron-2.3-blockRangeQuery-BBI
    )
    for branch in "${branches[@]}"; do
        {
            echo "Building images for $branch"
            buildImages "$branch"

            ./original-startFabric.sh go &> /dev/null
            sleep 10
            pushd ./go

            insert_first_1M
            run_range_tests_1M

            insert_remaining
            run_range_tests_10M

            popd
            ./original-networkDown.sh &> /dev/null
        } >> "$results" 2>&1
    done
}

run_range_tests_1M() {
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x000000000000d3b2c76221467d2f8c8f1de832a2 -r 100 -keylist "1M-versions.txt"
    done
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x03fb320c81ad2a55de600b13967879c341706afb -r 1000 -keylist "1M-versions.txt"
    done
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x0c52b252da5cc314c744b7715206282bf3ba9eb4 -r 5000 -keylist "1M-versions.txt"
    done
}

run_range_tests_10M() {
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x0000000000022da6024a5657b6a1d7f4fff03315 -r 100 -keylist "10M-versions.txt"
    done
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x0000005c7dc69d405f09aaadca29068d4f88cde8 -r 1000 -keylist "10M-versions.txt"
    done
    for i in {1..6}; do
        go run application.go -t getHistoryForAssetRange -k 0x0000ac61d2ebe805e9de54f430075f896d0afaac -r 5000 -keylist "10M-versions.txt"
    done
}

insert_first_1M() {
    local dataFile="/home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17000000-17010000.json"

    printf "Inserting %s\n\n" "$dataFile"
    go run application.go -t BulkInvokeParallel -f "$dataFile"
    printf "\n"
}

insert_remaining() {
    local filenames=(
        "blockTransactions17010001-17011000.json" 
        "blockTransactions17011001-17012000.json" 
        "blockTransactions17012001-17015000.json"
        "blockTransactions17015001-17020000.json"
        "blockTransactions17020001-17030000.json"
        "blockTransactions17030001-17050000.json"
        "blockTransactions17090001-17100000.json"
    )
    local dataDir="/home/andrey/Documents/insert-tpch/ethereum/First100K"

    for file in ${filenames[@]}; do
        echo "Inserting file: $dataDir/$file"
        go run application.go -t BulkInvokeParallel -f "$dataDir/$file"
        printf "\n"
    done
}

buildImages() {
    pushd /home/andrey/Desktop/fabric-rvp
    git checkout "$1"
    {
        make docker-clean 
        echo "y" | docker image prune
        make peer-docker
        make orderer-docker
    } &> /dev/null
    popd
}

main

exit 0
