#!/bin/bash
for i in {1..300};
# 7 versions * 300 = 2100 versions
# 10K * 300 = 3 000 000 records
do
    echo "running command $1"
    go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
done
