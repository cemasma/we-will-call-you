FROM golang:1.19-buster as builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -o ./main

FROM golang:1.19-buster as release
WORKDIR /app
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/repository/in_memory/en ./repository/in_memory/en
COPY --from=builder /app/repository/in_memory/tr ./repository/in_memory/tr
COPY --from=builder /app/main ./main
EXPOSE 8080
ENTRYPOINT ["./main"]