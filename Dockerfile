FROM bufbuild/buf:latest AS protobufs

COPY proto /proto
COPY buf.* /
RUN buf generate

FROM golang:1.25.0-alpine AS build

RUN apk add --no-cache build-base

COPY go.* /src/
COPY server /src/server
COPY --from=protobufs /gen /src/gen
RUN cd /src/server && CGO_ENABLED=1 go build

FROM alpine AS final

COPY --from=build /src/server/server /app/server
WORKDIR /app

CMD [ "./server" ]