#!/bin/bash

./original-startFabric.sh go

pushd go

go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/ethereum/First100K/blockTransactions17010001-17011000.json

go run application.go -t getHistoryForAsset -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4

go run application.go -t blockRangeQuery -start 10 -end 15 -u 3

go run application.go -t blockRangeQuery -start 10 -end 15 -u 100
