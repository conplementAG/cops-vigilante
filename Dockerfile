FROM golang:1.20.5-bullseye as build

RUN apt-get update && \
    apt-get install lsb-release -y

RUN go version

## Manage Go dependencies to allow caching #################
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

ADD . /app

WORKDIR /app/cmd/vigilante

RUN go build -o /out/vigilante

WORKDIR /app
RUN go vet ./...
RUN go test ./... --cover

FROM gcr.io/distroless/base:debug AS final

EXPOSE 8443

USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /out /app

WORKDIR /app
ENTRYPOINT ["./vigilante", "run", "--tls"]