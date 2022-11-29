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

        // Submit the specified transaction.
        // createCar transaction - requires 5 argument, ex: ('createCar', 'CAR12', 'Honda', 'Accord', 'Black', 'Tom')
        // changeCarOwner transaction - requires 2 args , ex: ('changeCarOwner', 'CAR12', 'Dave')
        // await contract.submitTransaction('createCar', 'CAR12', 'Honda', 'Accord', 'Black', 'Tom');
        //const orderData = {
        //    "L_ORDERKEY": 1,
        //    "L_PARTKEY": 155190,
        //    "L_SUPPKEY": 7706,
        //    "L_LINENUMBER": 1,
        //    "L_QUANTITY": 17,
        //    "L_EXTENDEDPRICE": 21168.23,
        //   "L_DISCOUNT": 0.04,
        //    "L_TAX": 0.02,
        //    "L_RETURNFLAG": "N",
        //    "L_LINESTATUS": "O",
        //    "L_SHIPDATE": "1996-03-13T07:00:00.000Z",
        //    "L_COMMITDATE": "1996-02-12T07:00:00.000Z",
        //    "L_RECEIPTDATE": "1996-03-22T07:00:00.000Z",
        //    "L_SHIPINSTRUCT": "DELIVER IN PERSON",
        //    "L_SHIPMODE": "TRUCK",
        //    "L_COMMENT": "egular courts above the"
        //};
        //const orderDataString = JSON.stringify(orderData);
        //console.info(orderDataString);
        //const orderBuffer = Buffer.from(orderDataString);
        
        //const jsonDataPath = path.resolve(__dirname, 'mnt', 'hgfs', 'term-project', 'firstElems.json');
        elapsedTime("Start json transactions");
        const jsonDataPath = '/mnt/hgfs/term-project/firstElems.json';
        //console.info(jsonDataPath);
        const jsonData = JSON.parse(fs.readFileSync(jsonDataPath, 'utf8'));
        //console.info(jsonData);
        //await contract.submitTransaction('addOrder', orderDataString);
        //await contract.submitTransaction('addOrder', 1, 1, 155190, 7706);
        //for(let i = 13; i < 513; i++) {
        //    await contract.submitTransaction('createCar', 'CAR' + i, 'Honda', 'Accord', 'Black', 'Owner' + i);
        //}
        let transactionList = [];
        
        const length = jsonData.table.length;
        for (let i = 0; i < length; i++) {
            let orderDataString = JSON.stringify(jsonData.table[i]);
            await contract.submitTransaction('addOrder', orderDataString);
            //transactionList.push(contract.submitTransaction('addOrder', orderDataString));
        }
        
        //await Promise.all(transactionList);
        
        elapsedTime("Finished json transactions"); 
        // The following are the fastest txn times as configtx.BatchTimeout has been set to 0.0001 seconds, and 1000 txns takes longer than 1s
        // for loop: 1000 records takes about 28.83 seconds => 1,000,000 will take approximately 8 hours
        // Promise.all: 1000 records takes about 10.514 seconds with await inside for loop => 1,000,000 will take approximately 2.92 hours
        // TODO: Test bulk contract method using Promise.all and sending all data in one packet to contract
        
        console.log('Transaction has been submitted');

        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
}

main();
