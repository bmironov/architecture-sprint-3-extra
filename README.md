# architecture-sprint-3-extra

## Introduction

To simplify development single "monolight" app was coded in Go language. Based
on this code single Docker image is produced. Based on supplied environment
variable values same image accesses different databses.


## Database configuration

Run folowing scripts against databases in specified containers
| script | DB's container |
| -------| -------------- |
| ```init_db_hvac.sql``` | ```postgres-hvac``` | ```psql postgresql://hvac_user:hvac_password@localhost:5433/warm_home_hvac``` |
| ```init_db_light.sql``` | ```postgres-light``` | ```psql postgresql://light_user:light_password@localhost:5434/warm_home_light``` |

This way each database will have own subset of all tables supported by the app.
It will make it quite easy to see if traffic is properly routed.

To access each app directly use following examples:
```
curl -s http://localhost:8080/hvac/1
curl -s http://localhost:8081/lights/1
```

## Building application image
```
docker build --tag warm-home .
или
docker compose up --build -d
```

## Configure Kong

### Start Kong DB
```
docker run -d --name postgres-kong \
  --network=warm_home \
  -p 5434:5432 \
  -e "POSTGRES_USER=kong" \
  -e "POSTGRES_DB=kong" \
  -e "POSTGRES_PASSWORD=kongpass" \
  postgres:alpine
```

### Initialize Kong DB
```
docker run --rm --network=warm-home \
 -e "KONG_DATABASE=postgres" \
 -e "KONG_PG_HOST=postgres-kong" \
 -e "KONG_PG_PORT=5432" \
 -e "KONG_PG_PASSWORD=kongpass" \
 -e "KONG_PASSWORD=test" \
kong/kong-gateway:latest kong migrations bootstrap
```

### Creating services
```
curl -i -s -X POST http://localhost:8001/services \
  --data name=warm-home-hvac \
  --data url='http://warm-home-hvac:8080'
curl -i -s -X POST http://localhost:8001/services \
  --data name=warm-home-light \
  --data url='http://warm-home-light:8080'
```

### Deleting service
```
curl -X DELETE http://localhost:8001/services/warm-home-hvac
curl -X DELETE http://localhost:8001/services/warm-home-light
```

### Viewing service configuration
```
curl -X GET http://localhost:8001/services/warm-home-hvac | jq
curl -X GET http://localhost:8001/services/warm-home-light | jq
```

### Listing services
```
curl -X GET http://localhost:8001/services | jq
```

### Creating routes
```
curl -i -X POST http://localhost:8001/services/warm-home-hvac/routes \
  --data 'paths[]=/hvac' \
  --data name=hvac-route
curl -i -X POST http://localhost:8001/services/warm-home-light/routes \
  --data 'paths[]=/lights' \
  --data name=light-route
```

### Deleting routes
```
curl -X DELETE http://localhost:8001/routes/hvac-route
curl -X DELETE http://localhost:8001/routes/light-route
```

### Viewing route configuration
```
curl -X GET http://localhost:8001/services/warm-home-hvac/routes/hvac-route | jq
curl -X GET http://localhost:8001/services/warm-home-light/routes/light-route | jq
```

### Listing routes
```
curl http://localhost:8001/routes | jq
```

### Test routing
It looks a bit ugly to repeat ```hvac``` or ```lights``` twice in URLs, but
point is to route traffic properly. Notice that we use port 8000 here, which is
provided by Kong API Gateway.

```
curl -s http://localhost:8000/hvac/hvac/1 | jq
curl -s http://localhost:8000/hvac/lights/1 | jq   # Should fail
curl -s http://localhost:8000/lights/lights/1 | jq
```

## Kafka

Initial idea was to create "monolith" application that supports all types of
REST API for HVAC and Lighting systems based on configuration provided via
environment variables. This way we produce single Docker image, but in effect
can start it twice and have two separate applications each supporting own type
of devices.

Next step is to "split" Lighting monolith into two parts: async processor of
lighting telemetry data (Kafka consumer) and Lighting microservice with the rest
of original Lighting monolith functionality. First part will stay mostrly the
same, but instead of calls to process telemetry calls from Lighting devices and
directly save them into PostgreSQL, it will produce new messages in related
Kafka topic. Then separate application in ```src/telemetry.lights``` will be
simple Kafka condumer that has only one purpose - receive messages from topic
and save them into PostgreSQL. At this point our ```docker-compose.yaml``` has 2
databases (one for HAVA, one for Lighting) and 3 applications (HVAC "monolith",
Lighting microservice, Lighting telemetry consumer).

### Auto topic creation
For simplicity of development, Kafka is configured to create new topics
automatically via ```KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"``` setting.

### Splitting 'lights' monolith

The 'Lights' monolith has been split into 2 parts:
1. Controller that was processing telemetry instead of inserting data into DB
   directly now commits this data into Kafka topic. This app is residing in
   ```warm_home_light``` container
2. New application in ```src/telemetry.lights``` has been added. It is simple
   Kafka consumer, that listens to topic and executes ```INSERT``` statements
   into ```lights_telemetry``` table. This app is residing in
   ```warm_home_light_telemetry``` container and is configured via environment
   variables.

Below is example of ```PUT``` request to publish fresh lights telemetry data:
```
PUT http://localhost:8000/lights/lights/1/telemetry
content-type: application/json

{
    "current_bright": 80,
    "target_bright": 90
}
```

To monitor async data processing ```docker logs -f <container_name>``` command
can be utilized against each of these two containers.

Logs from ```warm-home-light-telemetry``` will show messages like following:
```
Processing message light_id=1, current: 80.00 target: 90.00 ... new ID=44
```
Text before ellipses shows data from JSON that was retrieved from Kafka's topic.
Text after ellipses shows response from PostgreSQL in form of automatically
generated ID (see definition of the table where column ```light_telemetry_id```
is ```BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY```). This way we see both
sides of interaction between Kafka's consumer and PostgreSQL client.

To get list of last 100 telemetry datapoints use
```
GET http://localhost:8000/lights/lights/1/telemetry
```

## Minikube

Configure minikube to use Docker driver by default
```
minikube config set driver docker
```

Start
```
minikube start --memory=4096 --cpus=3
```

Access GUI
```
minikube dashboard --url
```

Set k8s namespace
```
#kubectl create namespace warm-home
#kubectl config set-context --current --namespace=warm-home
#kubectl config view | grep namespace:
```

Set minikube to use local Docker images:
```
eval $(minikube docker-env)
```

Build application images
```
docker build -f Dockerfile.web -t warm-home-hvac:latest .
docker build -f Dockerfile.web -t warm-home-light:latest .
docker build -f Dockerfile.telemetry.lights -t warm-home-light-telemetry:latest .
```

Convert ```docker-compose.yaml``` to k8s
```
cd k8s
rm *yaml
kompose -f ../docker-compose.yaml convert
```

Apply k8s YAML files
```
cd k8s
kubectl apply -f .
```

Check k8s services
```
minikube service --all
```

Describe k8s service and init Kong DB:
```
kubectl describe svc postgres-kong

docker run --rm \
 -e "KONG_DATABASE=postgres" \
 -e "KONG_PG_HOST=10.96.207.102" \
 -e "KONG_PG_PORT=5432" \
 -e "KONG_PG_PASSWORD=kongpass" \
 -e "KONG_PASSWORD=test" \
kong/kong-gateway:latest kong migrations bootstrap
```

Init HVAC & Lighting databases
```
# HVAC DB
POSTGRES_HVAC_POD=`kubectl get pods | grep postgres-hvac | awk '{print $1}'`
kubectl exec --stdin --tty $POSTGRES_HVAC_POD -- /bin/bash
psql postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}
# run commands from ./init_db_hvac.sql

# LIGHTING DB
POSTGRES_LIGHT_POD=`kubectl get pods | grep postgres-light | awk '{print $1}'`
kubectl exec --stdin --tty $POSTGRES_LIGHT_POD -- /bin/bash
psql postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}
# run commands from ./init_db_light.sql
```

Deploy single node Kafka from (Strimzi)[https://strimzi.io/quickstarts/]
```
kubectl create namespace kafka
kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
kubectl get pod -n kafka --watch
kubectl apply -f https://strimzi.io/examples/latest/kafka/kraft/kafka-single-node.yaml -n kafka 

# Test
kafka-producer -ti --image=quay.io/strimzi/kafka:0.44.0-kafka-3.8.0 --rm=true --restart=Never -- bin/kafka-console-producer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic my-topic
kubectl -n kafka run kafka-consumer -ti --image=quay.io/strimzi/kafka:0.44.0-kafka-3.8.0 --rm=true --restart=Never -- bin/kafka-console-consumer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic my-topic --from-beginning

kubectl get pods -o wide -n kafka
kubectl describe pod my-cluster-dual-role-0 -n kafka
```

Initialize Kong Gateway (we'll be using minikube's SSH tunnels to pods)
```
KONG_IP=`kubectl get svc kong-gateway|grep kong-gateway|awk '{print $3}'`
KONG_PORT=`ps -ef|grep ssh|grep $KONG_IP|awk -v port=8001 '{for(i=1; i<=NF; i++){if ($i ~ port){split($i,a,":");print a[1]}}}'`
APP_PORT=`ps -ef|grep ssh|grep $KONG_IP|awk -v port=8000 '{for(i=1; i<=NF; i++){if ($i ~ port){split($i,a,":");print a[1]}}}'`

curl -i -s -X POST http://localhost:$KONG_PORT/services \
  --data name=warm-home-hvac \
  --data url='http://warm-home-hvac.default:8080'
curl -i -s -X POST http://localhost:$KONG_PORT/services \
  --data name=warm-home-light \
  --data url='http://warm-home-light.default:8080'
curl -i -X POST http://localhost:$KONG_PORT/services/warm-home-hvac/routes \
  --data 'paths[]=/hvac' \
  --data name=hvac-route
curl -i -X POST http://localhost:$KONG_PORT/services/warm-home-light/routes \
  --data 'paths[]=/lights' \
  --data name=light-route

curl -X GET http://localhost:$KONG_PORT/services | jq
curl -X GET http://localhost:$KONG_PORT/routes | jq

curl -s http://localhost:$APP_PORT/hvac/hvac/1 | jq
curl -s http://localhost:$APP_PORT/hvac/lights/1 | jq   # Should fail
curl -s http://localhost:$APP_PORT/lights/lights/1 | jq

curl --request PUT --header "Content-Type: application/json" \
  --data '{"current_temp":22.0,"target_temp":22.5}' \
  http://localhost:$APP_PORT/hvac/hvac/1/telemetry | jq

curl -X GET http://localhost:$APP_PORT/hvac/hvac/1/telemetry | jq

curl --request PUT --header "Content-Type: application/json" \
  --data '{"current_bright":80.5,"target_bright":92.0}' \
  http://localhost:$APP_PORT/lights/lights/1/telemetry | jq

curl -X GET http://localhost:$APP_PORT/lights/lights/1/telemetry | jq
    
```