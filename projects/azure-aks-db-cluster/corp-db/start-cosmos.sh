#!/bin/bash


max=30
poll=5

start-cosmos(){

  AZURE_COSMOS_EMULATOR_PARTITION_COUNT=1 \
  docker run \
    --publish 8081:8081 \
    --publish 10250-10255:10250-10255 \
    --name cosmos-emulator \
    --detach \
    mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest

  while [[ $max -ne 0 ]]; do
    response=$(curl --insecure https://localhost:8081 --write-out "%{http_code}")
    if [[ $response -eq "200" ]]; then
      ((max=1))
    else
      sleep poll
    fi
    echo "waiting for cosmos emulator try $count out of $max"
  done

}

fetch-certificates(){
  echo "grabbing certificate from cosmos container"
  response=$(curl --insecure https://localhost:8081/_explorer/emulator.pem --write-out "%{http_code}" > ~/emulatorcert.crt)
  if [[ $response == "200" ]]; then
    echo "successfully grabbed certificate"
  else
    echo "could not grab certificate, exiting"
    exit 1
  fi

}

install-certificates(){
  cp ~/emulatorcert.crt /usr/local/share/ca-certificates/
  sudo update-ca-certificates
  sudo update-ca-trust
}


main() {
  start-cosmos
  wait
  fetch-certificates
  wait
  install-certificates
}
