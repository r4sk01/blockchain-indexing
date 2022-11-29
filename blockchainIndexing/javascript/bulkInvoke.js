/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');

let start = process.hrtime();
const elapsedTime = (note) => {
    const precision = 3;
    let elapsed = process.hrtime(start)[1] / 1000000; // divide by a million to get nano to milli
    console.log(process.hrtime(start)[0] + " s, " + elapsed.toFixed(precision) + " ms - " + note); // print message + time
    start = process.hrtime(); // reset the timer
};

async function main() {
    try {
        const argsList = process.argv;
        const fileUrl = argsList.length && argsList.length >= 3 && argsList[2];
        
        if (!fileUrl || !path.isAbsolute(fileUrl)) {
            console.error(`File URL must be provided and must be an absolute path`);
            console.info(`Usage: node bulkInvoke /path/to/JSON/file`);
            process.exit(1);
        }
        
        //const argsList = process.argv;
        // load the network configuration
        const ccpPath = path.resolve(__dirname, '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com', 'connection-org1.json');
        let ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const identity = await wallet.get('appUser');
        if (!identity) {
            console.log('An identity for the user "appUser" does not exist in the wallet');
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: 'appUser', discovery: { enabled: true, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('mychannel');

        // Get the contract from the network.
        const contract = network.getContract('blockchainIndexing');
        
        elapsedTime("Start json bulk transaction");
        //const jsonDataPath = '/mnt/hgfs/term-project/firstElems.json';

        //const jsonStringData = fs.readFileSync(jsonDataPath, 'utf8');
        //const jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        let jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        
        //const ordersBuffer = Buffer.from(jsonStringData);
        let ordersBuffer = Buffer.from(jsonStringData);
        console.info(Buffer.byteLength(ordersBuffer));
        
        //const jsonData = JSON.parse(jsonStringData);
        let jsonData = JSON.parse(jsonStringData);
        
        // TODO: Update this to increase data inserted
        //jsonData = { table: jsonData.table.slice(0, 5000) };
        //jsonStringData = JSON.stringify(jsonData);
        //ordersBuffer = Buffer.from(jsonStringData);
        
        const jsonDataLength = jsonData.table.length
        //const jsonDataLength = jsonData.length
        
        // Fabric grpc message length must be less than 100MB
        //const byteLimit = 100 * 1000000; // Txn gets killed at this limit
        //const byteLimit = 50 * 1000000; // Txn gets killed at this limit
        //const byteLimit = 25 * 1000000; // Txn gets killed at this limit
        const byteLimit = 10 * 1000000; // Txn gets killed at this limit
        const bufferSize = Buffer.byteLength(ordersBuffer);
        const reductionFactor = bufferSize / byteLimit;
        
        const splitLength = Math.floor(jsonDataLength / reductionFactor);
        
        // This pattern borrowed from https://stackoverflow.com/questions/8495687/split-array-into-chunks
        const splits = jsonData.table.reduce((resultArray, item, index) => { 
            const chunkIndex = Math.floor(index / splitLength);

            if(!resultArray[chunkIndex]) {
              resultArray[chunkIndex] = []; // start a new chunk
            }

            resultArray[chunkIndex].push(item);

            return resultArray;
        }, []);
        
        let transactionList = [];
        
        splits.forEach(async (memberArray) => {
            let memberString = JSON.stringify(memberArray);
            let memberBuffer = Buffer.from(memberString);
            
            //transactionList.push(contract.submitTransaction('addOrdersBulk', memberBuffer));
            await contract.submitTransaction('addOrdersBulk', memberBuffer);
        });
        
        //await Promise.all(transactionList);
        //await contract.submitTransaction('addOrdersBulk', ordersBuffer);
        
        elapsedTime("Finished json bulk transaction"); 
        // The following are the fastest txn times as configtx.BatchTimeout has been set to 0.0001 seconds, and 1000 txns takes longer than 1s
        // See addOrdersBulk method in /blockchain-indexing/chaincode/blockchainIndexing/javascript/blockchainIndexing.js for test results
        
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
