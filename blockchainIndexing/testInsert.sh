#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version & test refactored APIs
#
# Author: Daniel Garon
# Date: 2024-03-14

main() {
    local results=/home/andrey/Desktop/refactoringTest.txt
    local branches=(
        2.3-hlf-im-original
        2.3-hlf-im-version
        2.3-hlf-im-block
    )
    for branch in "${branches[@]}"; do
        {
            echo "Building images for $branch"
            buildImages "$branch"
            insert
            run_tests
        } >> "$results" 2>&1
    done
}

run_tests() {
    pushd ./go
    go run application.go -t GetHistoryForKey -k 0x00000000000124d994209fbb955e0217b5c2eca1
    go run application.go -t GetHistoryForKeyRange -k 0x00000000000124d994209fbb955e0217b5c2eca1
    go run application.go -t GetHistoryForVersionRange -k 0x00000000000124d994209fbb955e0217b5c2eca1 -s 3 -e 6
    go run application.go -t GetHistoryForBlockRange -s 10 -e 20 -u 3
    popd
}

insert() {
    local dataFile=/home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17010001-17011000.json
    ./startFabric.sh go &> /dev/null
    sleep 10
    pushd ./go
    printf "Inserting %s\n\n" "$dataFile"
    go run application.go -t BulkInvokeParallel -f "$dataFile"
    printf "\n"
    popd
    ./networkDown.sh &> /dev/null
    printf "\n"
}

buildImages() {
    pushd /home/andrey/Desktop/hlf-indexing-middleware
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