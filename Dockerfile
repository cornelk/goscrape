FROM golang:1.22.0 as base

ENV CGO_ENABLED 0

COPY . /build

WORKDIR /build

RUN go build -o goscrape .

FROM gcr.io/distroless/static-debian12

COPY --from=base /build/goscrape /

CMD ["/goscrape"]

