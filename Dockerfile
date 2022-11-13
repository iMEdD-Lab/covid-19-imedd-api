FROM golang:1.19-alpine AS build

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN apk update && apk add --no-cache make
RUN make

FROM alpine:latest

COPY --from=build /build/covid19-greece-api /usr/bin
COPY --from=build /build/internal/data/municipalities_ypes.csv /root/municipalities_ypes.csv
COPY --from=build /build/migrations /root/migrations

RUN apk add --no-cache
EXPOSE 8080
ENTRYPOINT ["covid19-greece-api"]
