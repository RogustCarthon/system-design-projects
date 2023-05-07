FROM golang:1.19
WORKDIR /app
COPY . .
RUN go mod download
# EXPOSE 8080
CMD ["go test -v -run=TestInsert"]
