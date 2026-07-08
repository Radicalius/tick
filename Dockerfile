FROM golang:1.25.0-alpine AS build

RUN apk add --no-cache build-base

COPY go.* /src/
COPY server /src/server
COPY gen /src/gen
RUN cd /src/server && CGO_ENABLED=1 go build

FROM alpine AS final

COPY --from=build /src/server/server /app/server
WORKDIR /app

RUN mkdir /data

CMD [ "./server" ]