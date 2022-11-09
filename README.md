# fabric-samples on FabricSharp

_Authors_: [Andrei Bachinin](https://github.com/r4sk01), [Nick Fabrizio](https://github.com/NFabrizio)

## Environment Set Up

_Best results for this application can be achieved by using Linux or a virtual machine running Linux._  

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
      1. In the terminal on your local environment, navigate to the directory where
         you want to clone the blockchain-indexing repo  
         `cd ~/path/to/your/directory`  
      2. In the terminal, run:  
         `git clone [clone-url-for-your-fork]`  
         The URL should be in the format git@github.com:YourUsername/blockchain-indexing.git  
3. Clone the FabricSharp repository to your local environment.  
   _If you already have the files downloaded to your local machine, skip to the next step._  
   1. In the terminal on your local environment, navigate to the directory where
      you want to clone the blockchain-indexing repo  
      `cd ~/path/to/your/directory`  
   2. In the terminal, run:  
      `git clone [clone-url]`  
      The URL should be in the format git@github.com:YourUsername/blockchain-indexing.git  
4. Modify the FabricSharp code.  
5. Build the FabricSharp images.  

_Python is required to run this application, and Python 3.8.9+ is highly recommended._  
_Pip version 22.0.3 is required to run this application._

1. Clone this repository to your local environment.  
   _If you already have the files downloaded to your local machine, skip to the next step._
2. Fork this Github repo.
   1. In a web browser, visit https://github.com/inf0rmatiker/model-service
   2. Click the Fork button in the upper right corner of the screen
   3. In the "Where should we fork this repository?" pop up, select your username.
      Github should create a fork of the repo in your account
3. Clone your fork of the model-service repo.
   1. In the terminal on your local environment, navigate to the directory where
      you want to clone the model-service repo  
      `cd ~/path/to/your/directory`
   2. In the terminal, run:  
      `git clone [clone-url-for-your-fork]`  
      The URL should be in the format git@github.com:YourUsername/model-service.git
4. Install the required Python packages in your Python environment.
5. In the terminal on your local environment, navigate to the directory where
   the model-service files are located.
6. In the terminal run the following command to install the required packages.  
   `pip3 install -r requirements.txt`  
   _If any errors are encountered while running this command, try upgrading your pip version using `pip install --upgrade pip`_
   
## Running the application




docker exec -it peer0.org1.example.com /bin/sh
control + C then control + D
docker cp ../data/mychannel peer0.org1.example.com:/var/hyperledger/production/ledgersData/chains/chains/
docker cp peer0.org2.example.com:/var/hyperledger/production/ledgersData/chains/chains/mychannel/ ../data/mychannel

