#!/bin/bash
#
# Purpose: Build images for each index version and insert 12M TPCH, outputs results to file
#
# Author: Daniel Garon
# Date: 2024-02-21
#

results=/home/andrey/Desktop/insertResults-TPCH-12M.txt

dataFile=/home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json

branches=(
    dgaron-2.3-blockRangeQueryOriginalIndex
    dgaron-2.3-blockRangeQuery-VBI
    dgaron-2.3-blockRangeQuery-BBI
)

main() {
    for branch in "${branches[@]}"; do
        echo "$branch"
        buildImages "$branch"
        insert
    done
}

insert() {
    printf "PARALLEL\n\n"
    for ((i = 0; i < 3; i++)); do
        ./startFabric.sh go &> /dev/null
        sleep 10
        pushd ./go || exit           
        printf "Inserting %s\n\n" "$dataFile"
        go run application.go -t BulkInvokeParallel -f "$dataFile"
        printf "\n"
        popd || exit
        ./networkDown.sh &> /dev/null
    done
    printf "\n"
}

buildImages() {
    pushd /home/andrey/Desktop/fabric-rvp || exit
    git checkout "$1"
    make docker-clean
    echo "y" | docker image prune
    make peer-docker
    make orderer-docker
    popd || exit
}

main >> "$results" 2>&1

exit 0