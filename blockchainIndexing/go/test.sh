#!/bin/bash

#go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/sortUnsort10500/unsortedMilEntries.json
go run application.go -t BulkInvokeParallel -f /home/andrey/Documents/insert-tpch/sortUnsort12KK/unsorted12KKEntries.json

#resultFile=results1M.txt
resultFile=results12M.txt

# Change block range from 100-200 to 500-1000

base="go run application.go -t"

commands=( 
# "pointQuery -v 1 -k 23488"
# "versionQuery -start 1 -end 7 -k 23488"
# "pointQueryOld -v 1 -k 91041"
# "versionQueryOld -start 1 -end 7 -k 23488"
"blockRangeQuery -start 500 -end 1000 -u 4"
)

> "$resultFile"

for command in "${commands[@]}"; do
	echo "go run application.go -t ${command[@]}" >> "$resultFile" 2>&1
    full_command=("${base[@]}" "${command[@]}")
	for (( i = 0; i < 6; i++ )); do
	 	eval "${full_command[@]}" >> "$resultFile" 2>&1
	done
    echo "" >> "$resultFile" 2>&1
done