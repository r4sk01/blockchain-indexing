'use strict';

const { Gateway, Wallets} = require('fabric-network');
const fs = require('fs');
const path = require('path');

const ccpPath = path.resolve(__dirname, '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com', 'connection-org1.json');
const walletPath = path.join(process.cwd(), 'wallet');
const argsList = process.argv;
const fileUrl = argsList.length && argsList.length >= 3 && argsList[2];

const CHUNK_LENGTH = 1000;
const SPLIT_LENGTH = 30;
const TIMEOUT_DURATION_MS = 30000


const elapsedTime = (note, reset = true) => {
    const precision = 3;
    let elapsed = process.hrtime(start)[1] / 1000000; // divide by a million to get nano to milli
    console.log(process.hrtime(start)[0] + " s, " + elapsed.toFixed(precision) + " ms - " + note); // print message + time
    if (reset) {
        start = process.hrtime(); // reset the timer
    }
};




async function main() {
    let start = process.hrtime();
    
    try {
        if (!fileUrl || !path.isAbsolute(fileUrl)) {
            console.error(`File URL must be provided and must be an absolute path`);
            console.info(`Usage: node bulkInvoke /path/to/JSON/file`);
            process.exit(1);
        }

        const cpp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        const identity = await wallet.get('appUser');
        if (!identity) {
            console.log('An identity for the user "appUser" does not exist in the wallet');
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        const gateway = new Gateway();
        await gateway.connect(ccp, {wallet, identity: 'appUser', discovery: { enabled: true, asLocalhost: true } });
        const network = await gateway.getNetwork('mychannel');
        const contract = network.getContract('blockchainIndexing');

        elapsedTime("Start json bulk chunk transaction", false);

        const jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        const jsonData = JSON.parse(jsonStringData);
        const chunkedData = chunk(jsonData.table, CHUNK_LENGTH);
        const splitData = chunk(chunkedData, SPLIT_LENGTH);

        for (const split of splitData) {
            console.log(`Running split index: ${splitData.indexOf(split)}`);
            for (const chunk of split) {
                const chunkBuffer = Buffer.from(JSON.stringify(chunk));
                try {
                    await contract.submitTransaction('addOrdersBulk', chunkBuffer);
                } catch (err) {
                    console.error(err)
                }
                await timeout(TIMEOUT_DURATION_MS);
            }
        }

        elapsedTime("Finished submitting bulk transaction", false); 
        console.log('Transaction has been submitted');

        await gateway.disconnect();
    } catch (error) {
        console.error(`Failed to submit transaction: ${error}`);
        process.exit(1);
    }
    
}

void main();
