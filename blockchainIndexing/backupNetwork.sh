#!/bin/bash

# Exit on first error
set -ex

DIRECTORY="network-backup"
PATH1="$DIRECTORY/peer1"
PATH2="$DIRECTORY/peer2"
PATH3="$DIRECTORY/orderer"
PATH4="$DIRECTORY/ordererOrgs"
PATH5="$DIRECTORY/peerOrgs"
PATH6="$DIRECTORY/artifacts"
PATH7="$DIRECTORY/genesis"
PATH8="$DIRECTORY/wallet"
PATH9="$DIRECTORY/fabric-ca"

PEER1="peer0.org1.example.com"
PEER2="peer0.org2.example.com"
ORDERER="orderer.example.com"

PEER_PROD_PATH="/var/hyperledger/production/"
ORDERER_PROD_PATH=$PEER_PROD_PATH"orderer/"
TEST_NET_PATH="../test-network"
ORGS_PATH="$TEST_NET_PATH/organizations"
FABRIC_CA_PATH="$ORGS_PATH/fabric-ca"
ORDERER_ORG_PATH="$ORGS_PATH/ordererOrganizations"
PEER_ORG_PATH="$ORGS_PATH/peerOrganizations"
ARTIFACTS_PATH="$TEST_NET_PATH/channel-artifacts"
GENESIS_PATH="$TEST_NET_PATH/system-genesis-block"

# Ensure that all directories exist before starting backup process
if [ ! -d "$DIRECTORY" ]; then
  echo "$DIRECTORY does not exist"
  mkdir $DIRECTORY
else
  echo "$DIRECTORY exists"
fi

if [ ! -d "$PATH1" ]; then
  echo "$PATH1 does not exist"
  mkdir $PATH1
else
  echo "$PATH1 exists"
fi

if [ ! -d "$PATH2" ]; then
  echo "$PATH2 does not exist"
  mkdir $PATH2
else
  echo "$PATH2 exists"
fi

if [ ! -d "$PATH3" ]; then
  echo "$PATH3 does not exist"
  mkdir $PATH3
else
  echo "$PATH3 exists"
fi

if [ ! -d "$PATH4" ]; then
  echo "$PATH4 does not exist"
  mkdir $PATH4
else
  echo "$PATH4 exists"
fi

if [ ! -d "$PATH5" ]; then
  echo "$PATH5 does not exist"
  mkdir $PATH5
else
  echo "$PATH5 exists"
fi

if [ ! -d "$PATH6" ]; then
  echo "$PATH6 does not exist"
  mkdir $PATH6
else
  echo "$PATH6 exists"
fi

if [ ! -d "$PATH7" ]; then
  echo "$PATH7 does not exist"
  mkdir $PATH7
else
  echo "$PATH7 exists"
fi

#if [ ! -d "$PATH8" ]; then
#  echo "$PATH8 does not exist"
#  mkdir $PATH8
#else
#  echo "$PATH8 exists"
#fi

if [ ! -d "$PATH9" ]; then
  echo "$PATH9 does not exist"
  mkdir $PATH9
else
  echo "$PATH9 exists"
fi

# Copy ledger files from Docker containers to backup directory
docker cp $PEER1:$PEER_PROD_PATH $PATH1
docker cp $PEER2:$PEER_PROD_PATH $PATH2
docker cp $ORDERER:$ORDERER_PROD_PATH $PATH3

# Copy files from test-network to backup directory
cp -r $ORDERER_ORG_PATH $PATH4
cp -r $PEER_ORG_PATH $PATH5
cp -r $ARTIFACTS_PATH $PATH6
cp -r $GENESIS_PATH $PATH7
cp -r $FABRIC_CA_PATH $PATH9


# Copy /wallet
#cp -r "./javascript/wallet" $PATH8
