#!/bin/bash

# imports  
. scripts/envVar.sh
. scripts/utils.sh

CHANNEL_NAME="$1"
DELAY="$2"
MAX_RETRY="$3"
VERBOSE="$4"
: ${CHANNEL_NAME:="mychannel"}
: ${DELAY:="3"}
: ${MAX_RETRY:="5"}
: ${VERBOSE:="false"}

if [ ! -d "channel-artifacts" ]; then
	mkdir channel-artifacts
fi

createChannelTx() {
	set -x
	configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/${CHANNEL_NAME}.tx -channelID $CHANNEL_NAME
	res=$?
	{ set +x; } 2>/dev/null
  verifyResult $res "Failed to generate channel configuration transaction..."
}

createChannel() {
	setGlobals 1
	# Poll in case the raft leader is not set yet
	local rc=1
	local COUNTER=1
	while [ $rc -ne 0 -a $COUNTER -lt $MAX_RETRY ] ; do
		sleep $DELAY
		set -x
		peer channel create -o localhost:7050 -c $CHANNEL_NAME --ordererTLSHostnameOverride orderer.example.com -f ./channel-artifacts/${CHANNEL_NAME}.tx --outputBlock $BLOCKFILE --tls --cafile $ORDERER_CA >&log.txt
		res=$?
		{ set +x; } 2>/dev/null
		let rc=$res
		COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
	verifyResult $res "Channel creation failed"
}

joinChannel() {
	FABRIC_CFG_PATH=$PWD/../config/
	ORG=$1
	PEER_PORT=$2
	PEERS=$3
	setGlobalsPeers $ORG $PEER_PORT
	local rc=1
	local COUNTER=1
	while [ $rc -ne 0 -a $COUNTER -lt $MAX_RETRY ] ; do
		sleep $DELAY
		set -x
		peer channel join -b $BLOCKFILE >&log.txt
		res=$?
		{ set +x; } 2>/dev/null
			let rc=$res
			COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
	verifyResult $res "After $MAX_RETRY attempts, peer${PEERS}.org${ORG} has failed to join channel '$CHANNEL_NAME' "
}

setAnchorPeer() {
  ORG=$1
  docker exec cli ./scripts/setAnchorPeer.sh $ORG $CHANNEL_NAME 
}

FABRIC_CFG_PATH=${PWD}/configtx

## Create channeltx
infoln "Generating channel create transaction '${CHANNEL_NAME}.tx'"
createChannelTx

FABRIC_CFG_PATH=$PWD/../config/
BLOCKFILE="./channel-artifacts/${CHANNEL_NAME}.block"

## Create channel
infoln "Creating channel ${CHANNEL_NAME}"
createChannel
successln "Channel '$CHANNEL_NAME' created"

cnt=0
current_port=7051
infoln "Joining org1 peers to the channel..."
while [ $cnt -lt 4 ]
do
	infoln "Joining org1 peer${cnt} at port ${current_port} to the channel..."
	joinChannel 1 $current_port $cnt
	current_port=$(($current_port+100))
	cnt=$(($cnt+1))
done

cnt=0
current_port=9051
infoln "Joining org2 peers to the channel..."
while [ $cnt -lt 4 ]
do
	infoln "Joining org2 peer${cnt} at port ${current_port} to the channel..."
	joinChannel 2 $current_port $cnt
	current_port=$(($current_port+100))
	cnt=$(($cnt+1))
done

cnt=0
current_port=10051
infoln "Joining org3 peers to the channel..."
while [ $cnt -lt 4 ]
do
	infoln "Joining org3 peer${cnt} at port ${current_port} to the channel..."
	joinChannel 3 $current_port $cnt
	current_port=$(($current_port+100))
	cnt=$(($cnt+1))
done

cnt=0
current_port=11051
infoln "Joining org4 peers to the channel..."
while [ $cnt -lt 4 ]
do
	infoln "Joining org4 peer${cnt} at port ${current_port} to the channel..."
	joinChannel 4 $current_port $cnt
	current_port=$(($current_port+100))
	cnt=$(($cnt+1))
done

## Set the anchor peers for each org in the channel
infoln "Setting anchor peer for org1..."
setAnchorPeer 1
infoln "Setting anchor peer for org2..."
setAnchorPeer 2
infoln "Setting anchor peer for org3..."
setAnchorPeer 3
infoln "Setting anchor peer for org4..."
setAnchorPeer 4

successln "Channel '$CHANNEL_NAME' joined"
