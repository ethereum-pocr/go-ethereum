# executed from a kubernetes probe for readiness
geth attach --exec '!eth.syncing'  /app/.ethereum/geth.ipc | grep true