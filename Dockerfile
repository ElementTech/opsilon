FROM golang:1.19-alpine

RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV PATH="${PATH}:/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin"

RUN go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" -o opsilon

CMD [ "./opsilon" ]
