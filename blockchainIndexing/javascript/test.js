/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const fs = require('fs');


async function main() {
    try {
        const argsList = process.argv;
        const fileUrl = argsList.length && argsList.length >= 3 && argsList[2];
        
        const jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        
        const jsonData = JSON.parse(jsonStringData);
        
        const shortenedData = jsonData.table.slice(0, 10000);
        console.info('shortenedData.length: ', shortenedData.length);
        
        const jsonDataLength = jsonData.table.length
        
        console.info('Initial array length: ', jsonDataLength);
        
        //const jsonStringData = fs.readFileSync(fileUrl, 'utf8');
        const ordersBuffer = Buffer.from(jsonStringData);
        console.info(Buffer.byteLength(ordersBuffer));
        
        // Fabric grpc message length must be less than 100MB
        const byteLimit = 100 * 1000000;
        const bufferSize = Buffer.byteLength(ordersBuffer);
        const reductionFactor = Math.ceil(bufferSize / byteLimit);
        console.info('reductionFactor: ', reductionFactor);
        
        const splitLength = Math.floor(jsonDataLength / reductionFactor);
        console.info('splitLength: ', splitLength);
        
        const splits = [];
        
        const result = jsonData.table.reduce((resultArray, item, index) => { 
            const chunkIndex = Math.floor(index / splitLength);

            if(!resultArray[chunkIndex]) {
              resultArray[chunkIndex] = []; // start a new chunk
            }

            resultArray[chunkIndex].push(item);

            return resultArray;
        }, []);
        
        console.info('result length: ', result.length);
        result.forEach((arrayItem, index) => {
            console.info('result item ' + index + ' length: ', arrayItem.length);
        });
        
        console.info(result[0][0]);
        console.info(result[4][0]);
        
        result.forEach((memberArray) => {
            let memberString = JSON.stringify(memberArray);
            let memberBuffer = Buffer.from(memberString);
            
            console.info('memberBuffer size', Buffer.byteLength(memberBuffer));
        });

    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
        process.exit(1);
    }
}

main();
