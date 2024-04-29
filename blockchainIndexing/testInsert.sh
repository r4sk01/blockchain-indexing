#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version & test refactored APIs
#
# Author: Daniel Garon
# Date: 2024-03-14

main() {
    results=/home/andrey/Desktop/branchTest.txt
    local branches=(
        2.3-hlf-im-original
        2.3-hlf-im-version
        2.3-hlf-im-block
    )
    for branch in "${branches[@]}"; do
        echo "Building images for $branch"
        buildImages "$branch"
        insert_and_test
    done
}

insert_and_test() {
    local dataFile=/home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17010001-17011000.json
    ./startFabric.sh go
    sleep 10
    pushd ./go
    {
        printf "Inserting %s\n\n" "$dataFile"
        go run application.go -t BulkInvokeParallel -f "$dataFile"
        printf "\n"
        go run application.go -t GetHistoryForKey -k 0x00000000000124d994209fbb955e0217b5c2eca1
        go run application.go -t GetHistoryForKeyRange -k 0x00000000000124d994209fbb955e0217b5c2eca1
        go run application.go -t GetHistoryForVersionRange -k 0x00000000000124d994209fbb955e0217b5c2eca1 -s 3 -e 6
        go run application.go -t GetHistoryForBlockRange -s 10 -e 20 -u 3
    } >> "$results" 2>&1
    popd
    ./networkDown.sh
    printf "\n"
}

buildImages() {
    pushd /home/andrey/Desktop/hlf-indexing-middleware
    git checkout "$1"
    make docker-clean 
    echo "y" | docker image prune
    make peer-docker
    make orderer-docker
    popd
}

main

exit 0