# company-service

This microservice handles company data.

Each company is defined by the following attributes:

* ID (UUID) - required
* Name (15 characters) - required, unique
* Description (3000 characters) - optional
* Amount of Employees (integer) - required
* Registered (boolean) - required
* Type (Corporation, NonProfit, Cooperative, Sole Proprietorship) - required

**Disclaimer**

This is **not** a production-ready service.

Several features are missing, including:

* Security-sensitive data (e.g., database passwords) stored in Docker secrets.
* Proper JWT authentication with a database of eligible users and roles.
* Non-root users in Docker containers.
* TLS support and proper proxy configuration.
* Proper database user privileges.

**## OpenAPI**

OpenAPI files for the service are located in the `openapi` folder.

* For the company microservice, see `openapi/company-service.yaml`.
* For the JWT token service, see `openapi/token.yaml`.

**## Setup**

**### Prerequisites**

You must have `docker` and `docker-compose` installed. Test cases use `postman`.

**## Set Environment Variables**

Create a `.env` file in the repository root with the database and JWT secret configuration:

```
DB_USER=user
DB_PASSWORD=password
JWT_SECRET=your-secret-key
```

Execute:

```sh
docker-compose up --build
```

You can check the created containers with:

```sh
docker ps | grep company
```

You can shut down the service with:

```sh
docker-compose down -v
```

Note: Verify that the volume is not left dangling with:

```sh
docker volume ls -f dangling=true
```

If `company-service_pgdata` is listed, remove it with:

```sh
docker volume rm company-service_pgdata
```

Once the containers are running, import the `Company Microservice API Tests.postman_collection.json` file into `postman`. Run the collection to view the results.

You can check the Kafka messages with:

```sh
docker exec -it company-service-kafka-1 kafka-console-consumer --bootstrap-server kafka:9092 --topic company_events --from-beginning
```
