FROM gcr.io/distroless/static-debian12

COPY goscrape /

CMD ["/goscrape"]

