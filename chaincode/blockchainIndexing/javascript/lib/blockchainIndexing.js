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
const FNV = require('fnv-plus');

function generateRandomUint32() {
    return crypto.randomBytes(4).readUInt32BE(0);
  }
  
function generateHashFuncs(numHashFuncs) {
  const hashFuncs = [];
  
  for (let i = 0; i < numHashFuncs; i++) {
    const seed1 = generateRandomUint32();
    const seed2 = generateRandomUint32();
  
    hashFuncs.push((data) => {
      const h1 = FNV.hash(data, 32, seed1).value;
      const h2 = FNV.hash(data, 32, seed2).value;
  
      return h1 ^ h2;
    });
  }
  
  return hashFuncs;
}
  
function isKeyPresentInBloomFilter(orderKey, bloomFilter) {
  // Number of hash functions
  const numHashFuncs = 22;
  const numBits = bloomFilter.length * 8;
  
  const hashFuncs = generateHashFuncs(numHashFuncs);
  
  for (const hashFunc of hashFuncs) {
    const index = hashFunc(orderKey) % numBits;
    const byteIndex = Math.floor(index / 8);
    const bitIndex = index % 8;
  
    if ((bloomFilter[byteIndex] & (1 << bitIndex)) === 0) {
      return false;
    }
  }
  
  return true;
}

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

    // This is a test method for trying to speed up bulk transaction processing time, but it needs more work from the client
    // end because there is still a limit to the payload (orderBuffer) size of 100MB
    async addOrdersBulkChunk(ctx, orderBuffer) {
        console.info('============= START : Add Orders Bulk Chunk ===========');
        
        const orders = orderBuffer.toString();
        const ordersObj = JSON.parse(orders);
        
        const chunkLength = 1000;
        const splitLength = 30;
        
        // This pattern borrowed from https://stackoverflow.com/questions/8495687/split-array-into-chunks
        const chunks = ordersObj.reduce((chunkArray, item, index) => {
            const chunkIndex = Math.floor(index / chunkLength);

            if(!chunkArray[chunkIndex]) {
              chunkArray[chunkIndex] = []; // start a new chunk
            }

            chunkArray[chunkIndex].push(item);

            return chunkArray;
        }, []);
        console.info(`chunks.length: ${chunks.length}`);
        
        const splits = chunks.reduce((resultArray, item, index) => {
            const splitIndex = Math.floor(index / splitLength);

            if(!resultArray[splitIndex]) {
              resultArray[splitIndex] = []; // start a new chunk
            }

            resultArray[splitIndex].push(item);

            return resultArray;
        }, []);
        console.info(`splits.length: ${splits.length}`);
        //console.info(`splits[0].length: ${splits[0].length}`);
        //console.info(`splits[1].length: ${splits[1].length}`);
        
        const timeoutDuration_ms = 15000;
        
        await splits.forEach(async (split, index) => {
            console.info(`Running split index: ${index}`);
            setTimeout(async () => {
                console.info(`setTimeout ran at ${index * timeoutDuration_ms}`);
                split.forEach(async memberArray => {
                    const length = ordersObj.length;
        
                    for (let i = 0; i < length; i++) {
                        const memberArrayObj = memberArray[i];
                        // const { L_ORDERKEY, L_LINENUMBER, ...orderRest } = memberArrayObj;
                        const { L_ORDERKEY, ...orderRest } = memberArrayObj;
        
                        // Fabric key must be a string
                        // const orderKey = L_ORDERKEY.toString() + '-' + L_LINENUMBER.toString();
                        const orderKey = L_ORDERKEY.toString();
                        const pacakagedOrder = {
                            docType: 'order',
                            ...orderRest
                        };
            
                        console.info('orderKey: ', orderKey);
                        //transactionList.push(ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder))));
                        await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
                    }
                })
                
                console.info(`setTimeout ${index} which started at ${index * timeoutDuration_ms} finished`);
            } , index * timeoutDuration_ms);
        });
        
        console.info('============= END : Add Orders Bulk Chunk ===========');
        return;
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
            
            //console.info('allResults: ', allResults);
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

  async getHistoryForKeyWithBloomFilter(ctx, orderKey) {
    let history = [];
  
    let blockNumber = await ctx.stub.getBlockchainInfo().then((info) => info.height - 1);
  
    while (blockNumber >= 0) {
      const blockBytes = await ctx.stub.getBlockByNumber(blockNumber);
      const block = BlockDecoder.decode(blockBytes);
  
      const bloomFilterBytes = block.metadata.metadata[block.metadata.metadata.length - 1];
      const bloomFilter = Buffer.from(bloomFilterBytes);
  
      if (isKeyPresentInBloomFilter(orderKey, bloomFilter)) {
        const resultsIterator = await ctx.stub.getHistoryForKey(orderKey);
        let result;
  
        for (const tx of block.data.data) {
          const txId = tx.payload.header.channel_header.tx_id;
  
          while (!(result = await resultsIterator.next()).done) {
            const { tx_id } = result.value;
            if (tx_id === txId) {
              history.push({
                value: result.value.value.toString('utf8'),
                timestamp: result.value.timestamp,
                txId: result.value.tx_id,
              });
            }
          }
  
          // Reset the iterator after using it
          await resultsIterator.close();
        }
      }
  
      blockNumber--;
    }
  
    return history;
  }

  async pointQueryBloom(ctx, orderKey, keyVersion) {
    const results = await this.getHistoryForKeyWithBloomFilter(ctx, orderKey);
    const sortedResults = results.sort((a, b) => a.timestamp.seconds - b.timestamp.seconds);
    const finResult = sortedResults[keyVersion];
    return JSON.stringify(finResult);
  }

  async versionQueryBloom(ctx, orderKey, keyVersionStart, keyVersionEnd) {
    const results = await this.getHistoryForKeyWithBloomFilter(ctx, orderKey);
    const sortedResults = results.sort((a, b) => a.timestamp.seconds - b.timestamp.seconds);
    let finResult = sortedResults.slice(keyVersionStart, keyVersionEnd);
    return JSON.stringify(finResult);
  }

  async queryOrderHistoryByRangeBloom(ctx, startKey, endKey) {
    const results = [];
    for await (const { key } of ctx.stub.getStateByRange(startKey, endKey)) {
      const history = await this.getHistoryForKeyWithBloomFilter(ctx, key);
      results.push({ key, history });
    }
    return JSON.stringify(results);
  }
  
}

// To re-deploy chaincode in Fabric, navigate to the test-network directory and run the following command
// ./network.sh deployCC -ccn blockchainIndexing -ccv 1 -ccs {integer incremented from previous version deployed} -cci NA -ccl javascript -ccp ../chaincode/blockchainIndexing/javascript

module.exports = BlockchainIndexing;
