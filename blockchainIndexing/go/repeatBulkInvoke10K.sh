#!/bin/bash
for i in {1..10}; 
do
    echo "running command $i"
    go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
done