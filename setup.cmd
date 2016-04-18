@echo off
REM PREPARE STEP 1

docker-machine create -d virtualbox proxy

FOR /f "tokens=*" %i IN ('docker-machine env proxy') DO %i
FOR /f "tokens=*" %i IN ('docker-machine ip proxy') DO set CONSUL_IP=%i
docker-compose -p setup -f docker-compose-setup.yml up -d consul

docker-machine create -d virtualbox --swarm --swarm-master --swarm-discovery="consul://%CONSUL_IP%:8500" --engine-opt="cluster-store=consul://%CONSUL_IP%:8500" --engine-opt="cluster-advertise=eth1:2376" swarm-master
docker-machine create -d virtualbox --swarm --swarm-discovery="consul://%CONSUL_IP%:8500" --engine-opt="cluster-store=consul://%CONSUL_IP%:8500" --engine-opt="cluster-advertise=eth1:2376" swarm-node-1
docker-machine create -d virtualbox --swarm --swarm-discovery="consul://%CONSUL_IP%:8500" --engine-opt="cluster-store=consul://%CONSUL_IP%:8500" --engine-opt="cluster-advertise=eth1:2376" swarm-node-2

FOR /f "tokens=*" %i IN ('docker-machine env swarm-master') DO %i
FOR /f "tokens=*" %i IN ('docker-machine ip swarm-master') DO set DOCKER_IP=%i
docker-compose -p setup -f docker-compose-setup.yml up -d registrator

FOR /f "tokens=*" %i IN ('docker-machine env swarm-node-1') DO %i
FOR /f "tokens=*" %i IN ('docker-machine ip swarm-node-1') DO set DOCKER_IP=%i
docker-compose -p setup -f docker-compose-setup.yml up -d registrator

FOR /f "tokens=*" %i IN ('docker-machine env swarm-node-2') DO %i
FOR /f "tokens=*" %i IN ('docker-machine ip swarm-node-2') DO set DOCKER_IP=%i
docker-compose -p setup -f docker-compose-setup.yml up -d registrator


REM PREPARE STEP 2
FOR /f "tokens=*" %i IN ('docker-machine ip proxy') DO set FLOW_PROXY_HOST=%i
set FLOW_CONSUL_ADDRESS=http://%CONSUL_IP%:8500
FOR /f "tokens=*" %i IN ('docker-machine env proxy') DO %i
set FLOW_PROXY_DOCKER_HOST=%DOCKER_HOST%
set FLOW_PROXY_DOCKER_CERT_PATH=%DOCKER_CERT_PATH%



REM PREPARE STEP 1 - single node

docker-machine create -d virtualbox flow
FOR /f "tokens=*" %i IN ('docker-machine env flow') DO %i
FOR /f "tokens=*" %i IN ('docker-machine ip flow') DO set CONSUL_IP=%i
FOR /f "tokens=*" %i IN ('docker-machine ip flow') DO set DOCKER_IP=%i
docker-compose -p setup -f docker-compose-setup.yml up -d consul
docker-compose -p setup -f docker-compose-setup.yml up -d registrator

REM PREPARE STEP 2  - single node
FOR /f "tokens=*" %i IN ('docker-machine ip flow') DO set FLOW_PROXY_HOST=%i
set FLOW_CONSUL_ADDRESS=http://%CONSUL_IP%:8500
set FLOW_PROXY_DOCKER_HOST=%DOCKER_HOST%
set FLOW_PROXY_DOCKER_CERT_PATH=%DOCKER_CERT_PATH%


REM STEP 3 DEPLOY
FOR /f "tokens=*" %i IN ('docker-machine env --swarm swarm-master') DO %i
docker-flow.exe --blue-green --target=app --service-path="/api/v1/books" --side-target=db --flow=deploy --flow=proxy

REM STEP 4 SCALING
docker-flow.exe --scale="+2" --flow=scale --flow=proxy


REM STEP 5 HAPROXY CFG
 docker cp docker-flow-proxy:/cfg/haproxy.cfg .