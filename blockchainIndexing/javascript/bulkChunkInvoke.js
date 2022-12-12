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

async function main() {
    try {
        const argsList = process.argv;
        const fileUrl = argsList.length && argsList.length >= 3 && argsList[2];
        
        if (!fileUrl || !path.isAbsolute(fileUrl)) {
            console.error(`File URL must be provided and must be an absolute path`);
            console.info(`Usage: node bulkInvoke /path/to/JSON/file`);
            process.exit(1);
        }
        
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
        
        elapsedTime("Start json bulk chunk transaction", false);
        
        let jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        let jsonData = JSON.parse(jsonStringData);
        
        const jsonDataLength = jsonData.table.length;
        
        // Number of records to include in each chunk
        const chunkLength = 1000;
        
        // Number of splits to create from groups of chunks (i.e., 30 groups of 1000 record chunks)
        const splitLength = 30;
        
        // This pattern borrowed from https://stackoverflow.com/questions/8495687/split-array-into-chunks
        // Reduce the array of data in jsonData to an unspecified number of arrays containing 1000, or chunkLength, records each
        const chunks = jsonData.table.reduce((chunkArray, item, index) => {
            const chunkIndex = Math.floor(index / chunkLength);

            if(!chunkArray[chunkIndex]) {
              chunkArray[chunkIndex] = []; // start a new chunk
            }

            // Push item into chunk array at this index
            chunkArray[chunkIndex].push(item);

            return chunkArray;
        }, []);
        
        // Reduce the arrays of chunks into an unspecified number of arrays containing 30, or splitLength, chunks each
        const splits = chunks.reduce((resultArray, item, index) => {
            const splitIndex = Math.floor(index / splitLength);

            if(!resultArray[splitIndex]) {
              resultArray[splitIndex] = []; // start a new chunk
            }

            resultArray[splitIndex].push(item);

            return resultArray;
        }, []);
        
        // Although we await on the transaction submission, it returns before the ledger has been fully updated,
        // so we have to set a time to wait before sending the next request otherwise the peers get overwhelmed
        // and start failing to handle transactions properly
        const timeoutDuration_ms = 30000;
        
        await splits.forEach(async (split, index) => {
            console.info(`Running split index: ${index}`);
            setTimeout(async () => {
                console.info(`setTimeout ran at ${index * timeoutDuration_ms}`);
                split.forEach(async memberArray => {
                    //console.info(`memberArray.length: ${memberArray.length}`);
                    let memberString = JSON.stringify(memberArray);
                    let memberBuffer = Buffer.from(memberString);
           
                    try {
                        await contract.submitTransaction('addOrdersBulk', memberBuffer);
                    } catch(err) {
                        console.error(err);
                    }
                })
                
                elapsedTime(`setTimeout ${index} which started at ${index * timeoutDuration_ms} finished`, false);
            } , index * timeoutDuration_ms);
        }); 
        
        /*
         * This method with 30 second timeout duration took about 17 minutes to insert all 1,000,000 records into ledger 
         * with a total of 771.9MB for ledger across 13 different block files
         */
        
        elapsedTime("Finished submitting bulk transaction", false); 
        
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
