FROM golang:1.20-alpine
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
COPY krm-functions/interface-fn/go.mod krm-functions/interface-fn/go.sum krm-functions/interface-fn/
RUN go mod download
RUN cd krm-functions/interface-fn && go mod download
COPY . .
RUN cd krm-functions/interface-fn; go build -o /usr/local/bin/function ./

FROM gcr.io/distroless/static:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]