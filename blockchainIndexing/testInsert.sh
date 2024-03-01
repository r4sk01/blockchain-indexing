#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version, insert 12M TPCH, & output results to file
#
# Author: Daniel Garon
# Date: 2024-02-21 
# Checked with shellcheck.net

main() {
    local results=/home/andrey/Desktop/insertResults-TPCH-12M.txt
    local branches=(
        dgaron-2.3-blockRangeQuery-OriginalIndex
        # dgaron-2.3-blockRangeQuery-VBI
        # dgaron-2.3-blockRangeQuery-BBI
    )
    for branch in "${branches[@]}"; do
        {
            echo "Building images for $branch"
            buildImages "$branch"
            for ((i = 0; i < 3; i++)); do
                insert
            done
        } >> "$results" 2>&1
    done
}

insert() {
    local dataFile=/home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json
    printf "SEQUENTIAL\n"
    ./startFabric.sh go &> /dev/null
    sleep 10
    pushd ./go
    printf "Inserting %s\n\n" "$dataFile"
    go run application.go -t BulkInvoke -f "$dataFile"
    printf "\n"
    popd
    ./networkDown.sh &> /dev/null
    printf "\n"
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
