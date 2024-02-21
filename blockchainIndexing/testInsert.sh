#!/bin/bash

results=/home/andrey/Desktop/insertResults-TPCH-12M.txt

dataFile=/home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json

function insert() {
    printf "PARALLEL\n\n" >> "$results"
    for ((i = 0; i < 3; i++)); do
        ./startFabric.sh go
        sleep 10
        pushd ./go || exit
        {
            printf "Inserting %s\n\n" "$dataFile"
            go run application.go -t BulkInvokeParallel -f "$dataFile"
            printf "\n"
        } >> "$results"
        popd || exit
        ./networkDown.sh
    done
    printf "\n" >> "$results"
}

function buildImages() {
    make docker-clean
    echo "y" | docker image prune
    make peer-docker
    make orderer-docker
}

pushd /home/andrey/Desktop/fabric-rvp || exit
git checkout dgaron-2.3-blockRangeQueryOriginalIndex
buildImages
popd || exit
printf "ORIGINAL\n\n" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp || exit
git checkout dgaron-2.3-blockRangeQuery-VBI
buildImages
popd || exit
printf "VERSION\n\n" >> "$results"
insert

pushd /home/andrey/Desktop/fabric-rvp || exit
git checkout dgaron-2.3-blockRangeQuery-BBI
buildImages
popd || exit
printf "BLOCK\n\n" >> "$results"
insert