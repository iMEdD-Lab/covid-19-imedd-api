# Covid-19 Data Greece API [New]

This is an API that provides data about COVID-19 pandemic in Greece.

It is developed by Sociality in collaboration with iMEdD.

### Data sources

By default, the application retrieves its data from these sources:

- [Cases per prefecture](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greece_cases_v2.csv)
- [Greece Covid-19 timeline](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greeceTimeline.csv)
- [Deaths per municipality](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/deaths%20covid%20greece%20municipality%2020%2021.csv)
- [Demographic information](https://github.com/Sandbird/covid19-Greece/blob/master/demography_total_details.csv)
- [Waste information](https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/viral_waste_water.csv)

Of course, other sources can be also used for feeding it, by changing the application's environment variables
(see below).

### Endpoints

Here is a short description of the application's endpoints. More detailed info is included in Swagger documentation
(see below).

#### Basic Endpoints

- `/deaths_per_municipality`: COVID-19 deaths per Greek municipality
- `/cases`: COVID-19 deaths per Greek prefecture
- `/timeline`: Gets full COVID-19 info for every date of a specific period
- `/demographics`: Gets full COVID-19 demographics info for every date and for a certain age category(0-17,18-39,40-64,65+)

#### Helper Endpoints

- `/health`: Just for a simple check if the application is up and running.
- `/timeline_fields`: Gets all filter fields for the `/timeline` endpoint

#### Geographical Endpoints

- `/regional_units`: Greece's prefectures geographical information
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

### Setup environment variables

Please export the variables below with values of your choice.

- `POSTGRES_USER`: database user
- `POSTGRES_PASSWORD`: database password

### Run using docker-compose

Docker-compose will start 3 containers. The API that you just built, the database needed for storing the data, and the
documentation container.

You can simply type:

```shell
docker-compose up
```

You can also add `-d` parameter to start all containers in the background.

- ***Also note that, if the application is running for the first time, all migrations and data hydration will take
  place. No more than 2 minutes are needed for this operation.***

### Healthcheck

```shell
http://localhost:8080/health
```

If this URL answers with `hello friend` message, congratulations! Your application is up and running!

### Enter database:

You can see the database contents by typing:

```shell
docker exec -it covid19-postgres psql -U ${POSTGRES_USER} -d covid19
```

### Other environment variables you may wish to change (if you know what you are doing!)

#### Data Sources

- `CASES_CSV_URL`: CSV file containing covid
  cases ([default](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greece_cases_v2.csv))
- `TIMELINE_CSV_URL`: CSV file containing greece COVID19 timeline
  info ([default](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/greeceTimeline.csv))
- `DEATHS_PER_MUNICIPALITY_CSV_URL`: CSV file containing deaths per greek
  municipality ([default](https://github.com/iMEdD-Lab/open-data/blob/master/COVID-19/deaths%20covid%20greece%20municipality%2020%2021.csv))
- `DEMOGRAPHICS_CSV_URL`: CSV file containing demographics information per date and per age category ([default](https://github.com/Sandbird/covid19-Greece/blob/master/demography_total_details.csv))
- `WASTE_CSV_URL`: CSV file containing waste information per week and year ([default](https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/viral_waste_water.csv))
- `YPES_MUNICIPALITIES_CSV_FILE`: CSV file containing municipalities together with their identification code, and populations of 2021 and 2022 (default file is `internal/data/municipalities_ypes.csv`)

Please keep in mind that if you want to change the data source files, you have to strictly follow their initial format.

#### Other env vars

- `POPULATE_DB`: Choose if the database will be populated with new data at startup and every 24h
- `PORT`: API port (default 8080)
- `MIGRATIONS_DIR`: Migrations directory

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

## Run locally without Docker

We assume that you have Go 1.19 installed. If not, check [here](https://go.dev/doc/install)

Start a dummy database by typing:

```shell
make db-start
```

You can stop it later on by typing `make db-stop`.

You can run the migrations for the database by typing:

```shell
make migrate
```

You can populate the db with data by typing:

```shell
make populate-db
```

You can enter the db by typing:

```shell
make db-login
```

### Test the application

```shell
make test
```