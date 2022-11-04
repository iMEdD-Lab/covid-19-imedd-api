# Covid-19 Data Greece API [New]

This is an API that provides data about COVID-19 pandemic in Greece.

It is developed by Sociality in collaboration with iMEdD.

### Data sources

By default, the application retrieves its data from these sources:

- [Cases per prefecture](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greece_cases_v2.csv)
- [Greece Covid-19 timeline](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greeceTimeline.csv)
- [Deaths per municipality](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/deaths%20covid%20greece%20municipality%2020%2021.csv)

Of course, other sources can be also used for feeding it, by changing the application's environment variables
(see below).

### Endpoints

Here is a short description of the application's endpoints. More detailed info is included in Swagger documentation
(see below).

#### Basic Endpoints

- `/deaths_per_municipality`: COVID-19 deaths per Greek municipality
- `/cases`: COVID-19 deaths per Greek county
- `/timeline`: Gets full COVID-19 info for every date of a specific period

#### Helper Endpoints

- `/health`: Just for a simple check if the application is up and running.
- `/timeline_fields`: Gets all filter fields for the `/timeline` endpoint

#### Geographical Endpoints

- `/counties`: Greece's county geographical information
- `/municipalities`: Greece's municipality geographical information

## How to run

We assume that you have Docker and Docker-Compose installed. If not,
check [here](https://docs.docker.com/engine/install/)
for Docker and [here](https://docker-docs.netlify.app/compose/install/) for Docker-Compose.

### Build the application container

You can build the application container by typing:

```shell
make container
```

and let docker do its magic.

### Run using docker-compose

Docker-compose will start 3 containers. The API that you just built, the database needed for storing the data, and the
documentation container.

You can simply type:

```shell
docker-compose up
```

You can also add `-d` parameter to start all containers in the background.

- ***Also note that, if the application is running for the first time, all migrations and data hydration will take
  place. Approximately 2 minutes are needed for this operation.***

### Healthcheck

```shell
http://localhost:8080/health
```

If this URL answers with `hello friend` message, congratulations! Your application is up and running!

## Rate Limiting

We introduced rate limiting set to **100 requests per minute**.

## Authentication

No authentication method is used for now.

## Documentation (Swagger)

If you want to only see the API's documentation without running the whole application via `docker-compose`, simply 
run:

```shell
make swagger-start
```

By opening with your browser `http://localhost:9000` you will see a full descriptive Swagger documentation.
Information about all the application endpoints available, etc.