#!/bin/bash

for (( i=0; i < 6; ++i )); do
    go run application.go -t blockRangeQuery -start 1000 -end 1500 -u 1000
done


# go run application.go -t versionQuery -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -start 100 -end 

# go run application.go -t getHistoryForAsset -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4

# go run application.go -t getHistoryForAssetRange -k 0xffec0067f5a79cff07527f63d83dd5462ccf8ba4 -r 3 -keylist 1M-versions.txt

# go run application.go -t blockRangeQuery -start 1000 -end 1200 -u 500
