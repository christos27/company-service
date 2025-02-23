# company-service

This is a microservice handling company data.

Each company is defined by the following attributes:
* ID (uuid) required
* Name (15 characters) required - unique
* Description (3000 characters) optional
* Amount of Employees (int) required
* Registered (boolean) required
* Type (Corporations | NonProfit | Cooperative | Sole Proprietorship) required

**Disclaimer**

This is **not** a production ready service.
There are plenty of things missing.
Some of them are:
* Security sensitive data, e.g. db passwords, saved in docker secrets
* Proper JWT Auth having a database with eligible users and roles
* Non root users at docker containers
* TLS support and proper Proxy configuration
* Proper DB user privileges

## OpenAPI

OpenAPI files of the service is provided under folder `openapi`.

For the company microservice check `openapi\company-service.yaml`.

For the JWT token service check `openapi\token.yaml`.

## Set up

### Prerequisites

You should have `docker` and `docker-compose` installed.
Test cases are using `postman`.

## Set env variables

Create a `.env` file on the root of the repository providing the needed configuration for the database and JWT secret:
```
DB_USER=user
DB_PASSWORD=password
JWT_SECRET=your-secret-key
```

Execute
```sh
docker-compose up --build
```
You can check the containers created with
```sh
docker ps | grep company
```

You can shut down the service with
```sh
docker-compose down -v
```

Note: You may check that the volume is not left dangling with
```sh
docker volume ls -f dangling=true
```

If `company-service_pgdata` is reported the remove it with
```sh
docker volume rm company-service_pgdata
```

Once the containers are up and running import to `postman` the file `Company Microservice API Tests.postman_collection.json`.
Run the collection to get the results.

You can check that kafka messages with
```sh
docker exec -it company-service-kafka-1 kafka-console-consumer --bootstrap-server kafka:9092 --topic company_events --from-beginning
```
