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
const elapsedTime = (note, reset = true) => {
    const precision = 3;
    let elapsed = process.hrtime(start)[1] / 1000000; // divide by a million to get nano to milli
    console.log(process.hrtime(start)[0] + " s, " + elapsed.toFixed(precision) + " ms - " + note); // print message + time
    if (reset) {
        start = process.hrtime(); // reset the timer
    }
};

/*
 * The purpose of this script is to attempt to speed up the bulk loading of transactions into the ledger.
 * 
 * This script is still under development. It does not work in its current state because there is no limit from the 
 * client on the size of the payload, but there is a limit on the payload size from the orderer and peers. When the payload
 * is more than 100MB, you will get a gRPC error of resource exhausted. In order to get this working, code needs to be added
 * to ensure the payload sent is less than 100MB.
 */

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
        
        elapsedTime("Start json bulk server chunk transaction");
        
        const jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        const jsonData = JSON.parse(jsonStringData);
        const jsonStringDataClean = JSON.stringify(jsonData.table);
        const jsonDataBuffer = Buffer.from(jsonStringDataClean);
        
        const byteLimit = 100 * 1000000;
        
        const jsonDataLength = jsonData.table.length;
        const bufferSize = Buffer.byteLength(jsonDataBuffer);
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
        
        await splits.forEach(async (memberArray) => {
            let memberString = JSON.stringify(memberArray);
            let memberBuffer = Buffer.from(memberString);
            
            try {
                await contract.submitTransaction('addOrdersBulkChunk', memberBuffer);
            } catch(err) {
                console.error(err);
            }
        });
        
        elapsedTime("Finished submitting bulk server chunk transaction"); 
        
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        // TODO: Determine why the gateway does not disconnect when using this method
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
