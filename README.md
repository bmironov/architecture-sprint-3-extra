# architecture-sprint-3-extra

## Introduction

To simplify development single "monolight" app was coded in Go language. Based
on this code single Docker image is produced. Based on supplied environment
variable values same image accesses different databses.


## Database configuration

Run folowing scripts against databases in specified containers
| script | DB's container |
| -------| -------------- |
| ```init_db_hvac.sql``` | ```postgres_hvac``` |
| ```init_db_light.sql``` | ```postgres_light``` |

This way each database will have own subset of all tables supported by the app.
It will make it quite easy to see if traffic is properly routed.

To access each app directly use following examples:
```
curl -s http://localhost:8080/hvac/1
curl -s http://localhost:8081/lights/1
```

## Building application image
```
docker build --tag warm_home .
или
docker compose up --build -d
```

## Configure Kong

### Start Kong DB
```
docker run -d --name postgres_kong \
  --network=warm_home \
  -p 5434:5432 \
  -e "POSTGRES_USER=kong" \
  -e "POSTGRES_DB=kong" \
  -e "POSTGRES_PASSWORD=kongpass" \
  postgres:latest
```

### Initialize Kong DB
```
docker run --rm --network=warm_home \
 -e "KONG_DATABASE=postgres" \
 -e "KONG_PG_HOST=postgres_kong" \
 -e "KONG_PG_PORT=5432" \
 -e "KONG_PG_PASSWORD=kongpass" \
 -e "KONG_PASSWORD=test" \
kong/kong-gateway:latest kong migrations bootstrap
```

### Creating services
```
curl -i -s -X POST http://localhost:8001/services \
  --data name=warm_home_hvac \
  --data url='http://warm_home_hvac:8080'
curl -i -s -X POST http://localhost:8001/services \
  --data name=warm_home_light \
  --data url='http://warm_home_light:8080'
```

### Deleting service
```
curl -X DELETE http://localhost:8001/services/warm_home_hvac
curl -X DELETE http://localhost:8001/services/warm_home_light
```

### Viewing service configuration
```
curl -X GET http://localhost:8001/services/warm_home_hvac | jq
```

### Listing services
```
curl -X GET http://localhost:8001/services | jq
```

### Creating routes
```
curl -i -X POST http://localhost:8001/services/warm_home_hvac/routes \
  --data 'paths[]=/hvac' \
  --data name=hvac_route
curl -i -X POST http://localhost:8001/services/warm_home_light/routes \
  --data 'paths[]=/lights' \
  --data name=light_route
```

### Deleting routes
```
curl -X DELETE http://localhost:8001/routes/hvac_route
curl -X DELETE http://localhost:8001/routes/light_route
```

### Viewing route configuration
```
curl -X GET http://localhost:8001/services/warm_home_hvac/routes/hvac_route | jq
curl -X GET http://localhost:8001/services/warm_home_light/routes/light_route | jq
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
curl -s http://localhost:8000/hvac/lights/1 | jq
curl -s http://localhost:8000/lights/lights/1 | jq
```

## Kafka

### Auto topic creation
```KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"```

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
