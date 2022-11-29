/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');

class BlockchainIndexing extends Contract {

    async initLedger(ctx) {
        console.info('============= START : Initialize Ledger ===========');

        console.info('============= END : Initialize Ledger ===========');
    }

    async addOrder(ctx, order) {
        console.info('============= START : Add Order ===========');
        
        const orderObj = JSON.parse(order);
        const { L_ORDERKEY, L_LINENUMBER, ...orderRest } = orderObj;
        
        // Fabric key must be a string
        const orderKey = L_ORDERKEY.toString() + '-' + L_LINENUMBER.toString();
        const pacakagedOrder = {
            docType: 'order',
            ...orderRest
        };
        
        await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
    }

    async addOrdersBulk(ctx, orderBuffer) {
        console.info('============= START : Add Orders Bulk ===========');
        
        const orders = orderBuffer.toString();
        const ordersObj = JSON.parse(orders);
        let transactionList = [];
        
        const length = ordersObj.length;
        
        for (let i = 0; i < length; i++) {
            const orderObj = ordersObj[i];
            const { L_ORDERKEY, L_LINENUMBER, ...orderRest } = orderObj;
        
            // Fabric key must be a string
            const orderKey = L_ORDERKEY.toString() + '-' + L_LINENUMBER.toString();
            const pacakagedOrder = {
                docType: 'order',
                ...orderRest
            };
            
            console.info('orderKey: ', orderKey);
            //transactionList.push(ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder))));
            await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
        }
        
        //await Promise.all(transactionList);
        
        // The following are from tests done to check txn times for putState in for loop and with Promise.all
        // for loop: 1000 records takes 1.4 seconds => 1,000,000 will take approximately 23.33 minutes
        // Promise.all: 1000 records takes 1.192 seconds => 1,000,000 will take approximately 19.87 minutes
        
        console.info('============= END : Add Orders Bulk ===========');
    }

    async queryOrdersByRange(ctx, startKey, endKey) {
        const allResults = [];
        
        for await (const {key, value} of ctx.stub.getStateByRange(startKey, endKey)) {
            const strValue = Buffer.from(value).toString('utf8');
            let record;
            try {
                record = JSON.parse(strValue);
            } catch (err) {
                console.log(err);
                record = strValue;
            }
            
            allResults.push({ Key: key, Record: record });
        }

        return JSON.stringify(allResults);
    }

    async queryAllOrders(ctx) {
        const startKey = '';
        const endKey = '';
        const allResults = [];
        for await (const {key, value} of ctx.stub.getStateByRange(startKey, endKey)) {
            const strValue = Buffer.from(value).toString('utf8');
            let record;
            try {
                record = JSON.parse(strValue);
            } catch (err) {
                console.log(err);
                record = strValue;
            }
            allResults.push({ Key: key, Record: record });
        }
        //console.info(allResults);
        return JSON.stringify(allResults);
    }

}

// To re-deploy chaincode in Fabric, navigate to the test-network directory and run the following command
// ./network.sh deployCC -ccn blockchainIndexing -ccv 1 -ccs {integer incremented from previous version deployed} -cci NA -ccl javascript -ccp ../chaincode/blockchainIndexing/javascript

module.exports = BlockchainIndexing;
