#!/bin/bash

results=insertResults-TPCH-12M.txt

dataFile=/home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json

function insert() {
    pushd /home/andrey/Documents/insert-tpch/blockchain-indexing/blockchainIndexing
    printf "PARALLEL\n\n" >> "$results"
    for ((i = 0; i < 3; i++)); do
        ./startFabric.sh go
        sleep 10
        pushd go
        printf "Inserting $dataFile\n\n" >> ../"$results"
        go run application.go -t BulkInvokeParallel -f "$dataFile" >> ../"$results" 2>&1
        popd
        ./networkDown.sh
    done
    printf "\n" >> "$results"
    popd
}

function buildImages() {
    make docker-clean
    echo "y" | docker image prune
    make peer-docker
    make orderer-docker
}

pushd /home/andrey/Desktop/fabric-rvp
git checkout dgaron-2.3-blockRangeQueryOriginalIndex
buildImages
popd
printf "ORIGINAL\n\n" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp
git checkout dgaron-2.3-blockRangeQuery-VBI
buildImages
popd
printf "VERSION\n\n" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp
git checkout dgaron-2.3-blockRangeQuery-BBI
buildImages
popd
printf "BLOCK\n\n" >> "$results"
insert