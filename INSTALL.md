# Prerequisites

You need to have docker installed.
The following were tested on a Windows machine running Rancher Desktop and Ubuntu 22.04 WSL.

# Set up database

Postgres db is used.
To build the image execute:
```sh
cd db
docker build -t companies-db .
```

After the image is build successful, run an instance with a persistent volume:
```sh
docker run --name company-db-1 -d -e POSTGRES_PASSWORD=<password> -p 5435:5432 -v <host/path/to/data>:/var/lib/postgresql/data companies-db
```
example:
```sh
 docker run --name company-db-1 -d -e POSTGRES_PASSWORD=password  -p 5435:5432 -v /home/chris/data:/var/lib/postgresql/data companies-db
```

If you don't need persistency run:
```sh
docker run --name company-db-1 -d -e POSTGRES_PASSWORD=<password> -p 5435:5432 companies-db
```

