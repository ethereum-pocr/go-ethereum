#!/bin/bash

echo $password > ~/.accountpassword

echo "################################### Start init geth node with genesis block #################################################"
geth init CustomGenesis.json
echo "################################### End init geth node with genesis block #################################################"

echo "################################### Starting geth miner node #################################################"
geth --mine --mine --miner.etherbase=0x6e45c195e12d7fe5e02059f15d59c2c976a9b730
