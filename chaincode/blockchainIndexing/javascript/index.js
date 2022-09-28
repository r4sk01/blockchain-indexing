/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const BlockchainIndexing = require('./lib/blockchainIndexing');

module.exports.BlockchainIndexing = BlockchainIndexing;
module.exports.contracts = [ BlockchainIndexing ];
