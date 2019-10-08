#!/usr/bin/env bash

function test(){
    eval "$@"
    [[ $? == 0 ]]
}

function initVanillaController(){
  VANILLA_USER=$(echo "$VANILLA_CONTROLLER" | sed "s|@.*||g")
  VANILLA_HOST=$(echo "$VANILLA_CONTROLLER" | sed "s|.*@||g")
  VANILLA_PORT=$(echo "$VANILLA_CONTROLLER" | cut -d':' -s -f2)
  VANILLA_PORT="${PORT:-22}"
}

function initMicroserviceFile() {
  echo "---
kind: iofog-microservice
spec:
  name: ${MICROSERVICE_NAME}
  agent:
    name: ${NAME}-0
    config:
      memorylimit: 8192
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  roothostaccess: false
  application: ${APPLICATION_NAME}
  volumes:
    - hostdestination: /tmp/microservice
      containerdestination: /tmp
      accessmode: rw
  ports:
    - internal: 443
      external: 5005
  env:
    - key: TEST
      value: 42
  routes:
    - ${MSVC1_NAME}
    - ${MSVC2_NAME}
  config:
    test_mode: true
    data_label: 'Anonymous_Person_2'" > test/conf/microservice.yaml
}

function initMicroserviceUpdateFile() {
  echo "---
kind: iofog-microservice
spec:
  name: ${MICROSERVICE_NAME}
  agent:
    name: ${NAME}-0
    config:
      memorylimit: 5555
      diskdirectory: /tmp/iofog-agent/
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  roothostaccess: false
  application: ${APPLICATION_NAME}
  volumes:
    - hostdestination: /tmp/updatedmicroservice
      containerdestination: /tmp
      accessmode: rw
  ports:
    - internal: 443
      external: 5443
    - internal: 80
      external: 5080
  env:
    - key: TEST
      value: 75
    - key: TEST_2
      value: 42
  routes:
    - ${MSVC1_NAME}
  config:
    test_mode: true
    test_data: 42
    data_label: 'Anonymous_Person_3'" > test/conf/updatedMicroservice.yaml
}

function initApplicationFiles() {
  APP="  name: $APPLICATION_NAME"
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
        config:
          bluetoothenabled: true # this will install the iofog/restblue microservice
          abstractedhardwareEnabled: false
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      roothostaccess: false
      volumes:
        - hostdestination: /tmp/msvc
          containerdestination: /tmp
          accessmode: z
      ports: []
      config:
        test_mode: true
        data_label: 'Anonymous_Person'
    # Simple JSON viewer for the heart rate output
    - name: $MSVC2_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate-ui:arm
        x86: edgeworx/healthcare-heart-rate-ui:x86
        registry: remote
      roothostaccess: false
      ports:
        # The ui will be listening on port 80 (internal).
        - external: 5000 # You will be able to access the ui on <AGENT_IP>:5000
          internal: 80 # The ui is listening on port 80. Do not edit this.
          publicmode: false # Do not edit this.
      volumes: []
      env:
        - key: BASE_URL
          value: http://localhost:8080/data"
  ROUTES="
    routes:
    # Use this section to configure route between microservices
    # Use microservice name
    - from: $MSVC1_NAME
      to: $MSVC2_NAME"

  echo -n "---
  kind: iofog-application
  spec:
  $APP" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
  echo "$ROUTES" >> test/conf/application.yaml
  echo -n "---
  kind: iofog-application
  spec:
    applications:
    - " > test/conf/root_application.yaml
  echo -n "$APP"| awk '{$1=$1};1' >> test/conf/root_application.yaml
  echo -n "$MSVCS" | awk '{print " ", $0}' >> test/conf/root_application.yaml
  echo "$ROUTES" | awk '{print " ", $0}' >> test/conf/root_application.yaml
}

function initLocalAgentFile() {
  echo "---
kind: iofog-agent
spec:
  agents:
    - name: ${NAME}-0
      image: ${AGENT_IMAGE}
      host: 127.0.0.1" > test/conf/local-agent.yaml
}

function initLocalControllerFile() {
    echo "---
kind: iofog-controlplane
spec:
  controlplane:
    images: 
      controller: ${CONTROLLER_IMAGE}
    iofoguser:
      name: Testing
      surname: Functional
      email: user@domain.com
      password: S5gYVgLEZV
    controllers:
    - name: $NAME
      host: 127.0.0.1
---
kind: iofog-connector
spec:
  connectors:
  - name: $NAME
    image: ${CONNECTOR_IMAGE}
    host: localhost" > test/conf/local.yaml
}

function initAgentsFile() {
  initAgents
  echo "---
  kind: iofog-agent
  spec:
    agents:" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    echo "  - name: $AGENT_NAME
      user: ${USERS[$IDX]}
      host: ${HOSTS[$IDX]}
      keyfile: $KEY_FILE" >> test/conf/agents.yaml
  done
}

function initAgents(){
  USERS=()
  HOSTS=()
  PORTS=()
  AGENT_NAMES=()
  AGENTS=($AGENT_LIST)
  for AGENT in "${AGENTS[@]}"; do
    local USER=$(echo "$AGENT" | sed "s|@.*||g")
    local HOST=$(echo "$AGENT" | sed "s|.*@||g")
    local PORT=$(echo "$AGENT" | cut -d':' -s -f2)
    local PORT="${PORT:-22}"

    USERS+=" "
    USERS+="$USER"
    HOSTS+=" "
    HOSTS+="$HOST"
    PORTS+=" "
    PORTS+="$PORT"
    AGENT_NAMES+=" "
    AGENT_NAMES+="$AGENT_NAME"
  done
  USERS=($USERS)
  HOSTS=($HOSTS)
  PORTS=($PORTS)
}

function checkController() {
  [[ "$NAME" == $(iofogctl -v -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe controller "$NAME" | grep "name: $NAME") ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe controlplane | grep "name: $NAME") ]]
}

function checkConnector() {
  [[ "$NAME" == $(iofogctl -v -n "$NS" get connectors | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe connector "$NAME" | grep "name: $NAME") ]]
}

function checkConnectors() {
  for CNCT in "$@"; do
    [[ "$CNCT" == $(iofogctl -v -n "$NS" get connectors | grep "$CNCT" | awk '{print $1}') ]]
    [[ ! -z $(iofogctl -v -n "$NS" describe connector "$CNCT" | grep "name: $CNCT") ]]
  done
}

function checkControllerNegative() {
  [[ "$NAME" != $(iofogctl -v -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkConnectorNegative() {
  [[ "$NAME" != $(iofogctl -v -n "$NS" get connectors | grep "$NAME" | awk '{print $1}') ]]
}

function checkMicroservice() {
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe microservice "$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  [[ "{\"data_label\":\"Anonymous_Person_2\",\"test_mode\":true}" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $4}') ]]
  [[ "memorylimit: 8192" == $(iofogctl -v -n "$NS" describe agent "${NAME}-0" | grep memorylimit | awk '{$1=$1};1' ) ]]
  # Check route
  [[ "$MSVC1_NAME, $MSVC2_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk -F '\t' '{print $6}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS" get microservices | grep "5005:443")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS" get microservices | grep "/tmp/microservice:/tmp")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS" describe microservice "$MICROSERVICE_NAME" -o "test/conf/msvc_output.yaml"
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "name: $MICROSERVICE_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5005") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- hostdestination: /tmp/microservice") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "containerdestination: /tmp") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:test") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:test-arm") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"42\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "data_label: Anonymous_Person_2") ]]
}

function checkUpdatedMicroservice() {
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe microservice "$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  [[ "{\"data_label\":\"Anonymous_Person_3\",\"test_data\":42,\"test_mode\":true}" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $4}') ]]
  [[ "memorylimit: 5555" == $(iofogctl -v -n "$NS" describe agent "${NAME}-0" | grep memorylimit | awk '{$1=$1};1' ) ]]
  [[ "diskdirectory: /tmp/iofog-agent/" == $(iofogctl -v -n "$NS" describe agent "${NAME}-0" | grep diskdirectory | awk '{$1=$1};1') ]]
  # Check route
  [[ "$MSVC1_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk -F '\t' '{print $6}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS" get microservices | grep "5443:443, 5080:80")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS" get microservices | grep "/tmp/updatedmicroservice:/tmp")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS" describe microservice "$MICROSERVICE_NAME" -o "test/conf/msvc_output.yaml"
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "name: $MICROSERVICE_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC1_NAME") ]]
  [[ -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5080") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 80") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- hostdestination: /tmp/updatedmicroservice") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "containerdestination: /tmp") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:test") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:test-arm") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"75\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST_2") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"42\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_data: 42") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "data_label: Anonymous_Person_3") ]]
}

function checkMicroserviceNegative() {
  [[ "$MICROSERVICE_NAME" != $(iofogctl -v -n "$NS" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
}

function checkApplication() {
  [[ "$APPLICATION_NAME" == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  [[ "$MSVC1_NAME," == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $3}') ]]
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $4}') ]]
  [[ "$MSVC1_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  # Check config
  [[ "{\"data_label\":\"Anonymous_Person\",\"test_mode\":true}" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $4}') ]]
  [[ "bluetoothenabled: true" == $(iofogctl -v -n "$NS" describe agent "${NAME}-0" | grep bluetooth | awk '{$1=$1};1' ) ]]
  # Check route
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $5}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS" get microservices | grep "5000:80")
  [[ "$MSVC2_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS" get microservices | grep "/tmp/msvc:/tmp")
  [[ "$MSVC1_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS" describe application "$APPLICATION_NAME" -o "test/conf/app_output.yaml"
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $APPLICATION_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- from: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "to: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "external: 5000") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- internal: 80") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- hostdestination: /tmp/msvc") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "containerdestination: /tmp") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:x86-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:arm-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- key: BASE_URL") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "value: http://localhost:8080/data") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "data_label: Anonymous_Person") ]]
}

function checkApplicationNegative() {
  [[ "$NAME" != $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ "$MSVC1_NAME" != $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  [[ "$MSVC2_NAME" != $(iofogctl -v -n "$NS" get microservices | grep "$MSVC2_NAME" | awk '{print $1}') ]]
}

function checkAgent() {
  AGENT_NAME=$1
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe agent "$AGENT_NAME" | grep "name: $AGENT_NAME") ]]
}

function checkAgentNegative() {
  AGENT_NAME=$1
  [[ "$AGENT_NAME" != $(iofogctl -v -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgents() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    checkAgent "$AGENT_NAME"
  done
}

function checkAgentsNegative() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    checkAgentNegative "$AGENT_NAME"
  done
}

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

function checkAgentListFromController() {
  local API_ENDPOINT=$(cat /tmp/api_endpoint.txt)
  local ACCESS_TOKEN=$(cat /tmp/access_token.txt)
  local LIST=$(curl --request GET \
--url $API_ENDPOINT/api/v3/iofog-list \
--header "Authorization: $ACCESS_TOKEN" \
--header 'Content-Type: application/json')
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    local UUID=$(echo $LIST | jq -r '.fogs[] | select(.name == "'"$AGENT_NAME"'") | .uuid')
    [[ ! -z "$UUID" ]]
  done
}