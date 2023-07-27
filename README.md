# Chaincode for HLF SAI 

_Authors_: [Andrei Bachinin](https://github.com/r4sk01), [Daniel Garon](https://github.com/dgaron)

## Environment Set Up

_Application was tested on Linux Ubuntu_  

0. In order to process, you will need to install several essential packages including docker, build-essentials, Go Language. For reference, you can navigate to [Hyperledger Fabric docs](https://hyperledger-fabric.readthedocs.io/en/release-2.2/prereqs.html).  
1. Build custom HLF Peer and Orderer images.  
   1. Make sure to clone the [fabric-rvp](https://github.com/dgaron/fabric-rvp) repository to your local machine.  
   2. Switch to the [dgaron-2.3-handlers](https://github.com/dgaron/fabric-rvp/tree/dgaron-2.3-handlers) branch.  
   3. In the terminal on your system, navigate to the directory that you cloned. Execute the following commands to build custom images:  
      `go mod tidy`  
      `make clean`  
      `make peer-docker`  
      `make orderer-docker`  
   4. To verify the result of previous step, you may execute:  
      `docker images`  
      Output of the previous command should have the following:  
      `hyperledger/fabric-orderer       2.3                              772d36cc6a59   16 hours ago    37.3MB`  
      `hyperledger/fabric-orderer       2.3.3                            772d36cc6a59   16 hours ago    37.3MB`  
      `hyperledger/fabric-orderer       amd64-2.3.3-snapshot-946ed1bab   772d36cc6a59   16 hours ago    37.3MB`  
      `hyperledger/fabric-orderer       latest                           772d36cc6a59   16 hours ago    37.3MB`  
      `<none>                           <none>                           dce59bae6f09   16 hours ago    638MB`  
      `hyperledger/fabric-peer          2.3                              f42ae35191cb   16 hours ago    56MB`  
      `hyperledger/fabric-peer          2.3.3                            f42ae35191cb   16 hours ago    56MB`  
      `hyperledger/fabric-peer          amd64-2.3.3-snapshot-946ed1bab   f42ae35191cb   16 hours ago    56MB`  
      `hyperledger/fabric-peer          latest                           f42ae35191cb   16 hours ago    56MB`  
      `<none>                           <none>                           da030ac0fa2b   16 hours ago    710MB`  
2. Substitute original HLF images with custom ones in network.  
   1. Make sure to clone the current [blockchain-indexing](https://github.com/r4sk01/blockchain-indexing) repository to your local machine.  
   2. Switch to the [ab-getHistoryForKeys](https://github.com/r4sk01/blockchain-indexing/tree/ab-getHistoryForKeys) branch.  
   3. Navigate to the root folder of the branch.  
   4. Open the `/test-network/docker/docker-compose-test-net.yaml` file.  
   5. HLF Network consists of 3 nodes: `orderer.example.com`, `peer0.org1.example.com`, `peer0.org2.example.com`. For Orderer the image section should look like the following `image: hyperledger/fabric-orderer:2.3.3`. For each Peer the image section should look like the following `image: hyperledger/fabric-peer:2.3.3`.  

## Running the application, Inserting the Data

1. In the terminal on your Linux VM, navigate to the directory where you cloned
   the blockchain-indexing repository and navigate to the blockchainIndexing
   directory within that repository.  
2. Start the network by running the following command:  
   `./startFabric.sh go`  
   _This should bring the network up with an orderer and two peers._  
3. In the terminal on your Linux VM, navigate to the /blockchainIndexing/go
   directory.  
   `cd go`  
4. There are two ways to insert the data.  
   1. Insert data Sequentially:  
      `go run application.go -t BulkInvoke -f <path to the data file>`  
   2. Example of sequential insertion:  
      `go run application.go -t BulkInvoke -f ~/Documents/insert-tpch/sortUnsort10500/unsorted100KEntries.json`  
   3. Insert data in Parallel:  
      `go run application.go -t BulkInvokeParallel -f <path to the data file>`  
   4. Example of parallel insertion:  
      `go run application.go -t BulkInvokeParallel -f ~/Documents/insert-tpch/sortUnsort10500/unsorted100KEntries.json`  

## Running the Queries

1. Range Query  
   1. Range Query that uses New Handler specific for proposed algorithm:  
      `go run application.go -t getHistoryForAssetRange -k 36643,36742`  
   2. Range Query that uses old traditional HLF tools for Range Query:  
      `go run application.go -t getHistoryForAssetRangeOld -k 36643,36742`  
2. GetHistoryForAsset - query that utilize getHistoryForKeys API that returns all the versions for given key:  
   `go run application.go -t getHistoryForAsset -k 36643`  
3. GetHistoryForAssets - query that utilize new getHistoryForKeys API that returns all the versions for given keys:  
   `go run application.go -t getHistoryForAssets -k 1,99,100,91041`  
4. GetHistoryForAssetsOld - query that utilize old traditional HLF getHistoryForKey multiple times to return all the versions for given keys:  
   `go run application.go -t getHistoryForAssetsOld -k 1,99,100,91041`  
5. Point Query  
   _Work In Progress!_  
6. Version Query  
   _Work In Progress!_  

## Bringing the Network Down  

1. In the terminal on your Linux VM, navigate to the directory where you cloned
   the blockchain-indexing repository and navigate to the blockchainIndexing
   directory within that repository.  
2. Bring the network down by running the following command:  
   `./networkDown.sh`  
   _This should bring the network down, shutting down the orderer and both peers as well as clear the ledger._  

## Useful Docker Commands  

1. Exec into a running Docker container.  
   `docker exec -it peer0.org1.example.com /bin/sh`  
2. Exit from a Docker exec session.  
   Press control + C, then press control + D.  
3. Check the length of blockchain from inside of the peer container.  
   `peer channel getinfo -c mychannel`  
