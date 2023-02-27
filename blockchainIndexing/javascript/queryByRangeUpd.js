/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const fs = require('fs');

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
 * The purpose of this script is to query for a range of values from the blockchain rather than the entire blockchain.
 *
 * This script is still under development. It is currently set up to use hard coded values as the range and has to be
 * updated manually to change the range. It should be relatively simple to update it to accept command line arguments
 * for the range.
 *
 * This script does not currently behave the way you might expect. Examples will help explain.
 * 1. The script does not always return all transactions in the given range. Sometimes, it only returns the last one
 *    in the specified range.
 * 2. If you request a range starting with 10000-1 and ending with 10010-1. This script will sometimes return transactions
 *    with keys of 100-1 or 1000-1.
 *
 * Since documentation for this framework is very sparse, more experimentation and exploration is needed to resolve/understand
 * these issues.
 */

async function main() {
    try {
        elapsedTime("Start queryByRangeUpd.js transaction", false);
        // load the network configuration
        const ccpPath = path.resolve(__dirname, '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com', 'connection-org1.json');
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

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

        // Evaluate the specified transaction.
        // queryCar transaction - requires 1 argument, ex: ('queryCar', 'CAR4')
        // queryAllCars transaction - requires no arguments, ex: ('queryAllCars')
        // const result = await contract.evaluateTransaction('queryAllCars');
        let finres = [];
        const startKey = 91041;
        const endKey = 91051;
        for (let key = startKey; key < endKey; key++){
            console.log(`key: ${key}`)
            let result = await contract.evaluateTransaction('queryOrderHistoryByKey', key);
            finres.push(result)
        }
        console.log(`Transaction has been evaluated, result is: ${finres.toString()}`);
        
        elapsedTime("queryByRangeUpd.js transaction is done", false);
        // Disconnect from the gateway.
        await gateway.disconnect();

    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
        process.exit(1);
    }
}

main();
