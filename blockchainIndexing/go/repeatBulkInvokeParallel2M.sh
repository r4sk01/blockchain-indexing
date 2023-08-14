#!/bin/bash
for i in {1..200};
# 7 versions * 200 = 1400 versions
# 10K * 200 = 2 000 000 records
do
    echo "running command $1"
    go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
done
