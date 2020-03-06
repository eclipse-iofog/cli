#!/usr/bin/env bash

function testDeployVolume(){
  DIR="/tmp/iofogctl_tests"
  initAgents
  echo "---
apiVersion: iofog.org/v1
kind: Volume
spec:
  source: $DIR
  destination: $DIR
  permissions: 666
  agents:
  - $NAME-0
  - $NAME-1" > test/conf/volume.yaml

  run mkdir $DIR
  for IDX in 1 2 3; do
    echo "test$IDX" > "$DIR/test$IDX"
  done
  run mkdir $DIR/testdir
  for IDX in 1 2 3; do
    echo "test$IDX" > "$DIR/testdir/test$IDX"
  done
  iofogctl -v -n "$NS" deploy -f test/conf/volume.yaml

  # Check files
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    for FILE_IDX in 1 2 3; do
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat /tmp/iofogctl_tests/test$FILE_IDX | grep "test$FILE_IDX"
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat /tmp/iofogctl_tests/testdir/test$FILE_IDX | grep "test$FILE_IDX"
    done
  done
}