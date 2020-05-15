# Node JSON Sender
An external node implementation for Proxeus core. Sends form data to a REST endpoint via POST request.

## Implementation

All form data as JSON to REST endpoint with HTTP auth headers set by env vars.

## Usage

It is recommended to start it using docker.

The latest image is available at `proxeus/node-json-sender:latest`

See the configuration paragraph for more information on what environments variables can be overridden

## Configuration

The following parameters can be set via environment variables. 


| Environmentvariable | Required | Default value
--- | --- |   --- |  
PROXEUS_INSTANCE_URL |  | http://127.0.0.1:1323
SERVICE_URL |  | http://localhost:SERVICE_PORT
SERVICE_PORT |  | 8015
SERVICE_SECRET |  | my secret
REGISTER_RETRY_INTERVAL |  | 5
JSON_SENDER_URL | X |
JSON_SENDER_CLIENT_ID | X |
JSON_SENDER_TENANT_ID | X |
JSON_SENDER_SECRET | X |
JSON_SENDER_OAUTH_URL | X |

## Deployment

The node is available as docker image and can be used within a typical Proxeus Platform setup by including the following docker-compose service:

```
version: '3.7'

networks:
  xes-platform-network:
    name: xes-platform-network

services:
  node-balance-retriever:
    image: proxeus/node-json-sender:latest
    container_name: xes_node-node-json-sender
    networks:
      - xes-platform-network
    restart: unless-stopped
    environment:
      PROXEUS_INSTANCE_URL: http://xes-platform:1323
      JSON_SENDER_URL: http://url:123/endpoint
      JSON_SENDER_CLIENT_ID: client_id
      JSON_SENDER_TENANT_ID: tenant_id
      JSON_SENDER_SECRET: secret
      JSON_SENDER_OAUTH_URL: oath_url
      SERVICE_SECRET: secret
      SERVICE_PORT: 8015
      SERVICE_URL: http://node-json-sender:8015
      TZ: Europe/Zurich
    ports:
      - "8015:8015"
```
