#!/usr/bin/env bash

function login() {
  local API_ENDPOINT="$1"
  local EMAIL="$2"
  local PASSWORD="$3"
  local LOGIN=$(curl --request POST \
--url $API_ENDPOINT/api/v3/user/login \
--header 'Content-Type: application/json' \
--data "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
  echo $LOGIN
  ACCESS_TOKEN=$(echo $LOGIN | jq -r .accessToken)
  [[ ! -z "$ACCESS_TOKEN" ]]
  echo "$ACCESS_TOKEN" > /tmp/access_token.txt
  echo "$API_ENDPOINT" > /tmp/api_endpoint.txt
}

function waitForMsvc() {
  ITER=0
  MS=$1
  NS=$2
  [ -z $3 ] && STATE="RUNNING" || STATE="$3" && echo $STATE

  while [ -z "$(iofogctl -n $NS get microservices | grep $MS | grep $STATE)" ] ; do
      ITER=$((ITER+1))
      # Allow for 300 sec so that the agent can pull the image
      if [ $ITER -gt 300 ]; then
          echo "Timed out. Waited $ITER seconds for $MS to be $STATE"
          exit 1
      fi
      sleep 1
  done
}

function waitForSvc() {
  NS="$1"
  SVC="$2"
  ITER=0
  EXT_IP=""
  while [ -z "$EXT_IP" ] && [ $ITER -lt 60 ]; do
      sleep 1
      EXT_IP=$(kctl get svc -n $NS | grep $SVC | awk '{print $4}')
      ITER=$((ITER+1))
  done
  # Error if empty
  [ ! -z "$EXT_IP" ]
  # Return via stdout
  echo "$EXT_IP"
}

function waitForProxyMsvc(){
  waitForSystemMsvc "iofog/proxy:latest" $1 $2 $3
}

function waitForSystemMsvc() {
  local NAME="$1"
  local HOST="$2"
  local USER="$3"
  local KEY_FILE="$4"
  local SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $KEY_FILE $USER@$HOST"

  echo "HOST=$HOST"
  echo "USER=$USER"
  echo "KEY_FILE=$KEY_FILE"
  echo "SSH_COMMAND=$SSH_COMMAND"

  ITER=0
  while [ -z "$($SSH_COMMAND -- sudo docker ps | grep $NAME)" ] ; do
      ITER=$((ITER+1))
      # Allow for 300 sec so that the agent can pull the image
      if [ "$ITER" -gt 300 ]; then
          echo "Timed out. Waited $ITER seconds for proxy to be running"
          exit 1
      fi
      sleep 1
  done
}