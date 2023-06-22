/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');
const shim = require('fabric-shim');
const { BlockDecoder } = require('fabric-client/lib/BlockDecoder');
const crypto = require('crypto');

class BlockchainIndexing extends Contract {

    async initLedger(ctx) {
        console.info('============= START : Initialize Ledger ===========');

        console.info('============= END : Initialize Ledger ===========');
    }

    async addOrder(ctx, order) {
        console.info('============= START : Add Order ===========');
        
        const orderObj = JSON.parse(order);
        // const { L_ORDERKEY, L_LINENUMBER, ...orderRest } = orderObj;
        const { L_ORDERKEY, ...orderRest } = orderObj;

        
        // Fabric key must be a string
        // const orderKey = L_ORDERKEY.toString() + '-' + L_LINENUMBER.toString();
        const orderKey = L_ORDERKEY.toString();
        const pacakagedOrder = {
            docType: 'order',
            ...orderRest
        };
        
        await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
    }


    // Handles bulk transactions sent as buffer
    async addOrdersBulk(ctx, orderBuffer) {
        console.info('============= START : Add Orders Bulk ===========');
        
        const orders = orderBuffer.toString();
        const ordersObj = JSON.parse(orders);
        let transactionList = [];
        
        const length = ordersObj.length;
        
        for (let i = 0; i < length; i++) {
            const orderObj = ordersObj[i];
            // const { L_ORDERKEY, L_LINENUMBER, ...orderRest } = orderObj;
            const { L_ORDERKEY, ...orderRest } = orderObj;
        
            // Fabric key must be a string
            // const orderKey = L_ORDERKEY.toString() + '-' + L_LINENUMBER.toString();
            const orderKey = L_ORDERKEY.toString();
            const pacakagedOrder = {
                docType: 'order',
                ...orderRest
            };
            
            console.info('orderKey: ', orderKey);
            await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
        }
        
        /* 
         * The following are from tests done to check txn times for putState in for loop and with Promise.all
         * for loop: 1000 records takes 1.4 seconds => 1,000,000 will take approximately 23.33 minutes
         * Promise.all: 1000 records takes 1.192 seconds => 1,000,000 will take approximately 19.87 minutes
         */
        
        console.info('============= END : Add Orders Bulk ===========');
        return;
    }

    async queryOrderHistoryByKey(ctx, orderKey) {

        const results = [];

        const iterator = await ctx.stub.getHistoryForKey(orderKey);
      
        while (true) {
          const result = await iterator.next();

          if (result.done) {
            break;
          }

          const assetValue = result.value.value.toString('utf8');
          let transactionId = result.value.txId;

          let asset = {
            value: assetValue,
            timestamp: result.value.timestamp,
            txId: transactionId
          };

          results.push(asset);
        }
        await iterator.close();
        return JSON.stringify(results);
      
    }
    
    async pointQuery(ctx, orderKey, keyVersion) {

        const results = [];

        const iterator = await ctx.stub.getHistoryForKey(orderKey);
      
        while (true) {
          const result = await iterator.next();

          if (result.done) {
            break;
          }

          const assetValue = result.value.value.toString('utf8');
          let transactionId = result.value.txId;

          let asset = {
            value: assetValue,
            timestamp: result.value.timestamp,
            txId: transactionId
          };

          results.push(asset);
        }
        await iterator.close();

        let sortedResults = results.sort((a, b) => a.timestamp.seconds - b.timestamp.seconds);
        let finResult = sortedResults[keyVersion];

        return JSON.stringify(finResult);
    }

    async versionQuery(ctx, orderKey, keyVersionStart, keyVersionEnd) {

        const results = [];

        const iterator = await ctx.stub.getHistoryForKey(orderKey);
      
        while (true) {
          const result = await iterator.next();

          if (result.done) {
            break;
          }

          const assetValue = result.value.value.toString('utf8');
          let transactionId = result.value.txId;

          let asset = {
            value: assetValue,
            timestamp: result.value.timestamp,
            txId: transactionId
          };

          results.push(asset);
        }
        await iterator.close();

        let sortedResults = results.sort((a, b) => a.timestamp.seconds - b.timestamp.seconds);
        let finResult = sortedResults.slice(keyVersionStart, keyVersionEnd);

        return JSON.stringify(finResult);
    }

    async queryOrderHistoryByRange(ctx, startKey, endKey) {
        const results = [];
        for await (const {key, value} of ctx.stub.getStateByRange(startKey, endKey)) {
            const iterator = await ctx.stub.getHistoryForKey(key);
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                const assetValue = result.value.value.toString('utf8');
                let transactionId = result.value.txId;
                let asset = {
                value: assetValue,
                timestamp: result.value.timestamp,
                txId: transactionId
                };
                results.push(asset);
            }
            await iterator.close();
        }
        return JSON.stringify(results);
    }
  
}

// To re-deploy chaincode in Fabric, navigate to the test-network directory and run the following command
// ./network.sh deployCC -ccn blockchainIndexing -ccv 1 -ccs {integer incremented from previous version deployed} -cci NA -ccl javascript -ccp ../chaincode/blockchainIndexing/javascript

module.exports = BlockchainIndexing;
