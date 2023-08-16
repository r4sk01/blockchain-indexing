#!/bin/bash
for i in {1..400};
# 7 versions * 400 = 2800 versions
# 10K * 400 = 4 000 000 records
do
    echo "running command $1"
    go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted10KEntries.json
done
