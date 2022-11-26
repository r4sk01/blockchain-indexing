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
//        console.info(ctx);
//        const cars = [
//            {
 //               color: 'blue',
//                make: 'Toyota',
//                model: 'Prius',
//                owner: 'Tomoko',
//            },
//            {
//                color: 'red',
 //               make: 'Ford',
//                model: 'Mustang',
 //               owner: 'Brad',
  //          },
   //         {
    //            color: 'green',
     //           make: 'Hyundai',
      //          model: 'Tucson',
       //         owner: 'Jin Soo',
        //    },
//            {
 //               color: 'yellow',
   //             make: 'Volkswagen',
     //           model: 'Passat',
       //         owner: 'Max',
         //   },
           // {
//                color: 'black',
  //            make: 'Tesla',
   //             model: 'S',
     //           owner: 'Adriana',
       //     },
         //   {
           //     color: 'purple',
             //   make: 'Peugeot',
               // model: 'owner',
//                205: 'Michel',
  //          },
    //        {
      //          color: 'white',
        //        make: 'Chery',
          //      model: 'S22L',
            //    owner: 'Aarav',
//            },
 //           {
   //             color: 'violet',
     //           make: 'Fiat',
       //         model: 'Punto',
         //       owner: 'Pari',
           // },
//            {
  //              color: 'indigo',
    //            make: 'Tata',
      //          model: 'Nano',
        //        owner: 'Valeria',
          //  },
//            {
  //              color: 'brown',
    //            make: 'Holden',
      //          model: 'Barina',
        //        owner: 'Shotaro',
          //  },
//        ];

//        for (let i = 0; i < cars.length; i++) {
//            cars[i].docType = 'car';
//            await ctx.stub.putState('CAR' + i, Buffer.from(JSON.stringify(cars[i])));
//            console.info('Added <--> ', cars[i]);
//        }

        console.info('============= END : Initialize Ledger ===========');
    }

    //async queryCar(ctx, carNumber) {
  //      const carAsBytes = await ctx.stub.getState(carNumber); // get the car from chaincode state
//        if (!carAsBytes || carAsBytes.length === 0) {
        //    throw new Error(`${carNumber} does not exist`);
      //  }
    //    console.log(carAsBytes.toString());
  //      return carAsBytes.toString();
//    }

    async addOrder(ctx, order) {
    //async addOrder(ctx, orderKey, lineNumber, partKey, suppKey) {
        console.info('============= START : Add Order ===========');
        
        // Fabric doesn't like an object being passed as an arg
        //const order = orderBuffer.toString();
        const orderObj = JSON.parse(order);
        const { L_ORDERKEY, ...orderRest } = orderObj;
        
        // Fabric key must be a string
        const orderKey = L_ORDERKEY.toString();
        const pacakagedOrder = {
            docType: 'order',
            ...orderRest
        };
        await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(pacakagedOrder)));
        //const order = {
        //    docType: 'order',
        //    lineNumber,
        //    partKey,
        //    suppKey
        //};
        
        //await ctx.stub.putState(orderKey, Buffer.from(JSON.stringify(order)));
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
        console.info(allResults);
        return JSON.stringify(allResults);
    }

//    async createCar(ctx, carNumber, make, model, color, owner) {
//        console.info('============= START : Create Car ===========');
//        console.info(ctx);

        

  //      const car = {
//            color,
            //docType: 'car',
          //  make,
        //    model,
      //      owner,
    //    };

  //      await ctx.stub.putState(carNumber, Buffer.from(JSON.stringify(car)));
//        console.info('============= END : Create Car ===========');
//    }

//    async queryAllCars(ctx) {
      //  const startKey = '';
    //    const endKey = '';
  //      const allResults = [];
//        for await (const {key, value} of ctx.stub.getStateByRange(startKey, endKey)) {
        //    const strValue = Buffer.from(value).toString('utf8');
      //      let record;
    //        try {
  //              record = JSON.parse(strValue);
//            } catch (err) {
              //  console.log(err);
            //    record = strValue;
          //  }
        //    allResults.push({ Key: key, Record: record });
      //  }
    //    console.info(allResults);
  //      return JSON.stringify(allResults);
//    }

  //  async changeCarOwner(ctx, carNumber, newOwner) {
//        console.info('============= START : changeCarOwner ===========');

    //    const carAsBytes = await ctx.stub.getState(carNumber); // get the car from chaincode state
  //      if (!carAsBytes || carAsBytes.length === 0) {
//            throw new Error(`${carNumber} does not exist`);
      //  }
    //    const car = JSON.parse(carAsBytes.toString());
  //      car.owner = newOwner;
//
//        await ctx.stub.putState(carNumber, Buffer.from(JSON.stringify(car)));
 //       console.info('============= END : changeCarOwner ===========');
//    }

}

// To re-deploy chaincode in Fabric, navigate to the test-network directory and run the following command
// ./network.sh deployCC -ccn blockchainIndexing -ccv 1 -ccs {integer incremented from previous version deployed} -cci NA -ccl javascript -ccp ../chaincode/blockchainIndexing/javascript

module.exports = BlockchainIndexing;
