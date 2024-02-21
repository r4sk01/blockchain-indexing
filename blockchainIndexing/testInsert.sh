#!/bin/bash

results=insertResults-TPCH-12M.txt

dataFile=/home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json

function insert() {
    pushd /home/andrey/Documents/insert-tpch/blockchain-indexing/blockchainIndexing
    echo "" >> "$results"
    echo "PARALLEL" >> "$results"
    for ((i = 0; i < 3; i++)); do
        ./startFabric.sh go
        sleep 10
        pushd go
        echo "Inserting $dataFile" >> ../"$results"
        go run application.go -t BulkInvokeParallel -f "$dataFile" >> ../"$results" 2>&1
        popd
        ./networkDown.sh
    done
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
echo "" >> "$results"
echo "ORIGINAL" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp
git checkout dgaron-2.3-blockRangeQuery-VBI
buildImages
popd
echo "" >> "$results"
echo "VERSION" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp
git checkout dgaron-2.3-blockRangeQuery-BBI
buildImages
popd
echo "" >> "$results"
echo "BLOCK" >> "$results"
insert