# use with docker-compose in smoke-tests directory
FROM golang:1.18 AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
RUN go mod download
RUN cd ./examples/ && go build -o /out/examples

FROM scratch AS bin
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/examples /app/examples
ENTRYPOINT ["/app/examples"]
