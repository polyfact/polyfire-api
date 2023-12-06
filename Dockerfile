FROM alpine:3

# We then install git and the package required to build llama and the go app
RUN apk add go make git g++ ffmpeg jq

# And we finally build the app
WORKDIR /go/src/github.com/polyfire/api

COPY . /go/src/github.com/polyfire/api/

RUN go get

RUN make build/server_start

ENTRYPOINT go run .
