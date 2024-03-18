#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version, insert data, run range query tests
#
# Author: Daniel Garon
# Date: 2024-03-18
# Checked with shellcheck.net

main() {
    local results=/home/andrey/Desktop/block_range_test.txt
    local branches=(
        2.3-hlf-im-original
        2.3-hlf-im-version
        2.3-hlf-im-block
    )
    pushd /home/andrey/Documents/insert-tpch/blockchain-indexing/blockchainIndexing
    for branch in "${branches[@]}"; do
        {
            echo "Building images for $branch"
            buildImages "$branch"

            # Ethereum test
            git checkout ethereum
            ./startFabric.sh go &> /dev/null
            sleep 10
            pushd ./go
            insert_ethereum_10M
            for i in {200..1000..100}; do
                run_test 100 "$i" 100 || break
            done
            popd
            ./networkDown.sh &> /dev/null

            # TPCH test
            git checkout tpch
            ./startFabric.sh go &> /dev/null
            sleep 10
            pushd ./go
            insert_tpch_12M
            for i in {200..1000..100}; do
                run_test 100 "$i" 7 || break
            done
            popd
            ./networkDown.sh &> /dev/null

        } >> "$results" 2>&1
    done
}

insert_ethereum_10M() {
    local filenames=(
        "blockTransactions17000000-17010000.json"
        "blockTransactions17010001-17011000.json" 
        "blockTransactions17011001-17012000.json" 
        "blockTransactions17012001-17015000.json"
        "blockTransactions17015001-17020000.json"
        "blockTransactions17020001-17030000.json"
        "blockTransactions17030001-17050000.json"
        "blockTransactions17090001-17100000.json"
    )
    local dataDir="/home/andrey/Documents/insert-tpch/ethereum/First100K"

    for file in "${filenames[@]}"; do
        echo "Inserting file: $dataDir/$file"
        go run application.go -t BulkInvokeParallel -f "$dataDir/$file"
        printf "\n"
    done
}

insert_tpch_12M() {
    local dataFile=/home/andrey/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
    printf "Inserting %s\n\n" "$dataFile"
    go run application.go -t BulkInvokeParallel -f "$dataFile"
    printf "\n"
}

run_test() {
    for _ in {1..6}; do
        go run application.go -t GetHistoryForBlockRange -s "$1" -e "$2" -u "$3"
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
