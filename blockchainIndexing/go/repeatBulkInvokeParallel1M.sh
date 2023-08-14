#!/bin/bash
for i in {1..100};
# 7 versions * 100 = 700 versions
# 10K * 100 = 1 000 000 records
do
    echo "running command $1"
    go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
done
