FROM golang:1.13.2-alpine AS builder

WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -installsuffix cgo

FROM scratch
COPY --from=builder /go/src/app/webdash .
COPY --from=builder /go/src/app/index.html .

ENV IN_CLUSTER=outside
ENV CLUSTER_CONFIG=/.kube/config
ENV CONN_PORT=8000

CMD [ "./webdash" ]
