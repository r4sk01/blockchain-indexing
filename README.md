# fabric-samples on FabricSharp

_Authors_: [Andrei Bachinin](https://github.com/r4sk01), [Nick Fabrizio](https://github.com/NFabrizio)

## Environment Set Up

_Best results for this application can be achieved by using Linux or a virtual machine (VM) running Linux and FabricSharp v2.2.0._  

1. Set up Linux virtual machine.  
   1. Download and install VmWare.  
   2. Download Ubuntu 22 ISO image.  
   3. Start VmWare and create new virtual machine.  
      During setup, allow at least 10GB of disk space and 5GB of RAM.  
2. Clone this repository to your local environment.  
   _If you already have the files downloaded to your local machine, skip to the next step._  
   1. Fork this Github repo.  
      1. In a web browser, visit https://github.com/r4sk01/blockchain-indexing  
      2. Click the Fork button in the upper right corner of the screen.  
      3. In the "Where should we fork this repository?" pop up, select your username.  
         Github should create a fork of the repo in your account  
   2. Clone your fork of the blockchain-indexing repo.  
      1. In the terminal on your Linux VM, navigate to the directory where
         you want to clone the blockchain-indexing repo  
         `cd ~/path/to/your/directory`  
      2. In the terminal on your Linux VM, run:  
         `git clone [clone-url-for-your-fork]`  
         The URL should be in the format git@github.com:YourUsername/blockchain-indexing.git  
3. Clone the [FabricSharp](https://github.com/ooibc88/FabricSharp) repository to
   your local environment.  
   _If you already have the files downloaded to your local machine, skip to the next step._  
   1. In the terminal on your Linux VM, navigate to the directory where
      you want to clone the FabricSharp repository  
      `cd ~/path/to/your/directory`  
   2. In the terminal on your Linux VM, run:  
      `git clone [clone-url]`  
      The URL should be in the format git@github.com:ooibc88/FabricSharp.git  
4. Modify the FabricSharp code.  
   _Before performing this step, you can try to skipping to the next step. If the code does not build, and you receive [this error](https://github.com/ooibc88/FabricSharp/issues/25), come back to this step._  
   1. On your Linux VM in a text editor, open the file in the FabricSharp
      repository at the path FabricSharp/images/peer/Dockerfile.  
   2. On line 28 of this Dockerfile, there should be a command `RUN apk update`.
      Change it to `RUN apk update --allow-untrusted`.  
   3. Make the same change on line 56 of the same Dockerfile.  
   4. Save the file changes.  
5. Build the FabricSharp images.  
   1. Follow the instructions in the FabricSharp README file for building the
      FabricSharp peer and orderer Docker images.  
   _This will create the Docker images on your local machine so that they can be used when running the application using the steps below._  
   _If the code does not build, and you receive [this error](https://github.com/ooibc88/FabricSharp/issues/25), ensure you have completed the step above to modify the FabricSharp code._  

## Running the application  

1. In the terminal on your Linux VM, navigate to the directory where you cloned
   the blockchain-indexing repository and navigate to the blockchainIndexing
   directory within that repository.  
2. Start the network by running the following command:  
   `./startFabric.sh javascript`  
   _This should bring the network up with an orderer and two peers._
3. In the terminal on your Linux VM, navigate to the /blockchainIndexing/javascript
   directory.  
   `cd javascript`
4. Install the Node modules with the following command:  
   `npm install`  
5. Enroll the Fabric network administrator.  
   `node enrollAdmin`  
6. Register a Fabric user.  
   `node registerUser`
7. Verify that the nodes in the network are running FabricSharp.  
   `docker logs peer0.org1.example.com`  
   _This will output the logs from peer0 in your terminal. At the top of the output, you should see SHARP (2.2.0)._
   _If the output from the logs matches the expected result, the network is running on FabricSharp._  
   
## Bringing the Network Down  

1. In the terminal on your Linux VM, navigate to the directory where you cloned
   the blockchain-indexing repository and navigate to the blockchainIndexing
   directory within that repository.  
2. Bring the network down by running the following command:  
   `./networkDown.sh`  
   _This should bring the network down, shutting down the orderer and both peers._  
   
## Useful Application Commands  

1. Enroll shortcut command  
   1. After starting the network, run the following command:  
      `cd javascript/ && node enrollAdmin.js && node registerUser.js && cd ..`  
2. Bulk load data into the blockchain ledger.  
   1. Ensure the data you want to import exists on or is accessible from your VM.  
      _Since the data set is large, it may make more sense to set up a shared directory between your VM and your computer. Use the VMWare System Settings -> Sharing option._  
   2. Navigate into the /blockchain-indexing/blockchainIndexing/javascript directory.  
   3. Run the following command:  
      `node bulkChunkInvoke.js /path/to/data/directory/fileName.json`  
      _/path/to/data/directory/fileName.json should be the path to the data you want to upload in JSON format._  
      _If you set up a shared directory between your VM and your computer, it should be accessible from your VM at /mnt/hgfs/your-directory-name_  
      * This command take about 17 minutes to complete when loading 1 million records on a VM with 4GB of memory and 2 processors.  
      * The 1 million records in the TPC-H data set use about 770MB of space, so ensure your VM has at least 2GB of available disk space before using this command.  
      * If you see any errors returned while running this command (e.g., peer timeouts), try increasing the timeoutDuration_ms value in bulkChunkInvoke.js.  

## Useful Docker Commands  

1. Exec into a running Docker container.  
   `docker exec -it peer0.org1.example.com /bin/sh`  
2. Exit from a Docker exec session.  
   Press control + C, then press control + D.  
3. Copy FabricSharp ledger files from a docker container to your Linux VM.  
   `docker cp peer0.org2.example.com:/var/hyperledger/production/ledgersData/chains/chains/mychannel/ ../data/mychannel`  
4. Copy FabricSharp ledger files from your Linux VM to a docker container.  
   `docker cp ../data/mychannel peer0.org1.example.com:/var/hyperledger/production/ledgersData/chains/chains/`  
