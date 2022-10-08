FROM golang:1.19-alpine AS build
COPY . /build
RUN apk update && apk add --no-cache make git
WORKDIR /build
RUN make

FROM alpine:latest

COPY --from=build /build/covid19-greece-api /usr/bin

RUN apk add --no-cache bash tmux
EXPOSE 8080
ENTRYPOINT ["covid19-greece-api"]
