#!/bin/bash

# Exit on first error
set -ex

DIRECTORY="network-backup"
PATH1="$DIRECTORY/peer1/production"
PATH2="$DIRECTORY/peer2/production"
PATH3="$DIRECTORY/orderer/orderer"
PATH4="$DIRECTORY/ordererOrgs/ordererOrganizations"
PATH5="$DIRECTORY/peerOrgs/peerOrganizations"
PATH6="$DIRECTORY/artifacts/channel-artifacts"
PATH7="$DIRECTORY/genesis/system-genesis-block"
PATH8="$DIRECTORY/wallets"

PEER1="peer0.org1.example.com"
PEER2="peer0.org2.example.com"
ORDERER="orderer.example.com"

PEER_PROD_PATH="/var/hyperledger/"
ORDERER_PROD_PATH=$PEER_PROD_PATH"production/"
TEST_NET_PATH="../test-network"
ORGS_PATH="$TEST_NET_PATH/organizations"
ORDERER_ORG_PATH="$ORGS_PATH/ordererOrganizations"
PEER_ORG_PATH="$ORGS_PATH/peerOrganizations"
ARTIFACTS_PATH="$TEST_NET_PATH/channel-artifacts"
GENESIS_PATH="$TEST_NET_PATH/system-genesis-block"

# Ensure that all directories exist before starting backup process
if [ ! -d "$DIRECTORY" ]; then
  echo "$DIRECTORY does not exist"
  exit 0
else
  echo "$DIRECTORY exists"
fi

if [ ! -d "$PATH1" ]; then
  echo "$PATH1 does not exist"
  exit 0
else
  echo "$PATH1 exists"
fi

if [ ! -d "$PATH2" ]; then
  echo "$PATH2 does not exist"
  exit 0
else
  echo "$PATH2 exists"
fi

if [ ! -d "$PATH3" ]; then
  echo "$PATH3 does not exist"
  exit 0
else
  echo "$PATH3 exists"
fi

# Copy ledger files to Docker containers
docker cp $PATH1 $PEER1:$PEER_PROD_PATH
docker cp $PATH2 $PEER2:$PEER_PROD_PATH
docker cp $PATH3 $ORDERER:$ORDERER_PROD_PATH

# Copy backup files to test-network directories
cp -r $PATH4 $ORGS_PATH
cp -r $PATH5 $ORGS_PATH
cp -r $PATH6 $TEST_NET_PATH
cp -r $PATH7 $TEST_NET_PATH


# Still getting these errors when querying after running this script
# Install /wallet ?


#2022-11-16T04:27:09.793Z - error: [ServiceEndpoint]: Error: Failed to connect before the deadline on Endorser- name: peer0.org1.example.com, url:grpcs://localhost:7051, connected:false, connectAttempted:true
#2022-11-16T04:27:09.794Z - error: [ServiceEndpoint]: waitForReady - Failed to connect to remote gRPC server peer0.org1.example.com url:grpcs://localhost:7051 timeout:3000
#2022-11-16T04:27:09.797Z - info: [NetworkConfig]: buildPeer - Unable to connect to the endorser peer0.org1.example.com due to Error: Failed to connect before the deadline on Endorser- name: peer0.org1.example.com, url:grpcs://localhost:7051, connected:false, connectAttempted:true
#    at checkState (/home/nfab/Documents/hyperledger-fabric/blockchain-indexing/blockchainIndexing/javascript/node_modules/@grpc/grpc-js/build/src/client.js:74:26)
#    at Timeout._onTimeout (/home/nfab/Documents/hyperledger-fabric/blockchain-indexing/blockchainIndexing/javascript/node_modules/@grpc/grpc-js/build/src/channel.js:500:17)
#    at listOnTimeout (internal/timers.js:554:17)
#    at processTimers (internal/timers.js:497:7) {
#  connectFailed: true
#}
#2022-11-16T04:27:12.830Z - error: [ServiceEndpoint]: Error: Failed to connect before the deadline on Discoverer- name: peer0.org1.example.com, url:grpcs://localhost:7051, connected:false, connectAttempted:true
#2022-11-16T04:27:12.830Z - error: [ServiceEndpoint]: waitForReady - Failed to connect to remote gRPC server peer0.org1.example.com url:grpcs://localhost:7051 timeout:3000
#2022-11-16T04:27:12.830Z - error: [ServiceEndpoint]: ServiceEndpoint grpcs://localhost:7051 reset connection failed :: Error: Failed to connect before the deadline on Discoverer- name: peer0.org1.example.com, url:grpcs://localhost:7051, connected:false, connectAttempted:true
#2022-11-16T04:27:12.831Z - error: [DiscoveryService]: send[mychannel] - no discovery results
#Failed to evaluate transaction: Error: DiscoveryService has failed to return results

