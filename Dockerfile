FROM golang:1.19-alpine

RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" -o opsilon

CMD [ "./opsilon" ]
