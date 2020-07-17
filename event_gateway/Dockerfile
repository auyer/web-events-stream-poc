# build stage
FROM golang:alpine AS build-env


RUN apk update && apk add --no-cache git
WORKDIR /src
COPY go.sum .
COPY go.mod .
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o goapp

# final stage
FROM scratch
COPY --from=build-env /src/goapp /
COPY --from=build-env /src/asset /asset
ENTRYPOINT ["./goapp"]