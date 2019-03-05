FROM golang:alpine AS build-env

# Install OS-level dependencies.
RUN apk add --no-cache curl git build-base

# Copy our source code into the container.
WORKDIR /application

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
# Install our golang dependencies and compile our binary.
RUN GOOS=linux go build -ldflags "-X main.version=%1" -a -o ./bin/app .

FROM alpine:3.9
WORKDIR /application

COPY --from=build-env /application/bin/app /application/bin/

EXPOSE 5001

ENTRYPOINT ["/application/bin/app service"]
