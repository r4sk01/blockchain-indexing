#!/bin/bash
#
# Purpose: Build images for each index version, insert 12M TPCH, & output results to file
#
# Author: Daniel Garon
# Date: 2024-02-21 
# Checked with shellcheck.net

main() {
    local results=/home/andrey/Desktop/insertResults-ethereum-sequential.txt
    local branches=(
        dgaron-2.3-blockRangeQuery-OriginalIndex
        # dgaron-2.3-blockRangeQuery-VBI
        # dgaron-2.3-blockRangeQuery-BBI
    )
    for branch in "${branches[@]}"; do
        {
            echo "Building images for $branch"
            buildImages "$branch"
            insert
        } >> "$results" 2>&1
    done
}

insert() {
    local dataDir="/home/andrey/Documents/insert-tpch/ethereum/First100K"
    printf "SEQUENTIAL\n"

    ./original-startFabric.sh go &> /dev/null
    sleep 10
    pushd ./go || exit

    for file in "$dataDir"/*; do
        printf "Inserting %s\n\n" "$file"
        go run application.go -t BulkInvoke -f "$file"
        printf "\n"
    done

    popd || exit
    ./original-networkDown.sh &> /dev/null

    printf "\n"
}

buildImages() {
    pushd /home/andrey/Desktop/fabric-rvp || exit
    git checkout "$1"
    {
        make docker-clean 
        echo "y" | docker image prune
        make peer-docker
        make orderer-docker
    } &> /dev/null
    popd || exit
}

main

exit 0
