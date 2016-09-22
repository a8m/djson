FROM golang:1.7

RUN go get github.com/Jeffail/gabs
RUN go get github.com/bitly/go-simplejson
RUN go get github.com/antonholmquist/jason
RUN go get github.com/mreiferson/go-ujson
RUN go get github.com/ugorji/go/codec

WORKDIR /go/src/github.com/a8m/djson
ADD . /go/src/github.com/a8m/djson
