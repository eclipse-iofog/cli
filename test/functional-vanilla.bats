#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"
NS2="$NS"_2
USER_PW="S5gYVgLEZV"
USER_EMAIL="user@domain.com"

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy vanilla Controller" {
  initVanillaController
  echo "---
apiVersion: iofog.org/v1
kind: ControlPlane
metadata:
  name: func-controlplane
spec:
  controllers:
  - name: $NAME
    ssh:
      user: $VANILLA_USER
      host: $VANILLA_HOST
      port: $VANILLA_PORT
      keyFile: $KEY_FILE
    version: $VANILLA_VERSION
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
---
apiVersion: iofog.org/v1
kind: Connector
metadata:
  name: $NAME
spec:
  ssh:
    user: $VANILLA_USER
    host: $VANILLA_HOST
    port: $VANILLA_PORT
    keyFile: $KEY_FILE
  version: $VANILLA_VERSION" > test/conf/vanilla.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/vanilla.yaml
  checkController
  checkConnector
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Connector legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy connector "$NAME" status
  checkLegacyConnector
}

@test "Get Controller logs after vanilla deploy" {
  test iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Agent legacy commands" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy application and test deploy idempotence" {
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Connect in another namespace using file" {
  test iofogctl -v -n "$NS2" connect -f test/conf/vanilla.yaml
  checkController "$NS2"
  checkConnector "$NS2"
  checkAgents "$NS2"
  checkApplication "$NS2"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS2" legacy agent "$AGENT_NAME" status
  done
}

@test "Disconnect other namespace" {
  test iofogctl -v -n "$NS2" disconnect
  checkControllerNegative "$NS2"
  checkConnectorNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}

@test "Connect in other namespace using flags" {
  initVanillaController
  CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  test iofogctl -v -n "$NS2" connect --name "$NAME" --endpoint "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
  checkConnector
  checkAgents
}

@test "Disconnect other namespace" {
  test iofogctl -v -n "$NS2" disconnect
  checkControllerNegative "$NS2"
  checkConnectorNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -v -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Delete namespaces" {
  test iofogctl delete namespace "$NS"
  test iofogctl delete namespace "$NS2"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}