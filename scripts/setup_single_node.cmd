@echo off
REM PREPARE STEP 1 - single node
FOR /f "tokens=*" %%i IN ('docker-machine env flow') DO %%i
FOR /f "tokens=*" %%i IN ('docker-machine ip flow') DO set CONSUL_IP=%%i
FOR /f "tokens=*" %%i IN ('docker-machine ip flow') DO set DOCKER_IP=%%i

docker-compose -p setup -f docker-compose-setup.yml up -d consul
docker-compose -p setup -f docker-compose-setup.yml up -d registrator

REM PREPARE STEP 2  - single node
FOR /f "tokens=*" %%i IN ('docker-machine ip flow') DO set FLOW_PROXY_HOST=%%i
set FLOW_CONSUL_ADDRESS=http://%CONSUL_IP%:8500
set FLOW_PROXY_DOCKER_HOST=%DOCKER_HOST%
set FLOW_PROXY_DOCKER_CERT_PATH=%DOCKER_CERT_PATH%