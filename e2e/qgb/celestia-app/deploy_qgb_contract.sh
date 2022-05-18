#!/bin/bash

# This script deploys the QGB contract and outputs the address in `qgb_address.txt` file

# check whether to deploy a new contract or no need
if [[ "${DEPLOY_NEW_CONTRACT}" != "true" ]]
then
  echo "no need to deploy a new QGB contract. exiting..."
  exit 0
fi

# check if environment variables are set
if [[ -z "${EVM_CHAIN_ID}" || -z "${PRIVATE_KEY}" ]] || \
   [[ -z "${TENDERMINT_RPC}" || -z "${CELESTIA_GRPC}" ]] || \
   [[ -z "${EVM_ENDPOINT}" ]]
then
  echo "Environment not setup correctly. Please set:"
  echo "EVM_CHAIN_ID, PRIVATE_KEY, TENDERMINT_RPC, CELESTIA_GRPC, EVM_ENDPOINT variables"
  exit -1
fi

# install needed dependencies
apk add curl

# wait for the node to get up and running
while true
do
  status_code=$(curl --write-out '%{http_code}' --silent --output /dev/null core0:26657/status) # TODO don't use a hardcoded address
  if [[ "$status_code" -eq 200 ]] ; then
    break
  fi
  echo "Waiting for node to be up..."
  sleep 5s
done

echo "deploying QGB contract..."

/bin/celestia-appd deploy \
  -x ${EVM_CHAIN_ID} \
  -d ${PRIVATE_KEY} \
  -t ${TENDERMINT_RPC} \
  -c ${CELESTIA_GRPC} \
  -z ${EVM_CHAIN_ID} \
  -e ${EVM_ENDPOINT} > /opt/output

echo $(cat /opt/output)

cat /opt/output | tail -n 1 | cut -d\  -f 4 > /opt/qgb_address.txt
