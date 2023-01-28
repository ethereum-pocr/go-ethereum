#!/bin/sh
cd /app

BOOTNODES=""
for node in $(cat /app/bootnodes.json)
do 
  BOOTNODES="$node,$BOOTNODES"
done


if [ ! -f initialized ]; then
    geth init --datadir .ethereum /app/genesis.json
    # mv .ethereum/geth/nodekey .keystore/nodekey
    echo initialized > initialized
fi
networkid=$(grep chainId genesis.json | grep -Eo '[0-9]+')
public_ip=$(/sbin/ip route|awk '/default/ { print $3 }')
SYNCMODE=snap
geth --networkid $networkid --datadir .ethereum --bootnodes $BOOTNODES --syncmode $SYNCMODE --http --http.addr=0.0.0.0 --http.port=8545 --http.api=web3,eth,net,clique --http.corsdomain=* --http.vhosts=* --ws --ws.addr=0.0.0.0 --ws.port=8546 --ws.api=web3,eth,net,clique --ws.origins=* --nat extip:$public_ip
