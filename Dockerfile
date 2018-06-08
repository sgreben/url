FROM golang:1.9.4-alpine3.7
RUN apk add --no-cache make
WORKDIR /go/src/github.com/sgreben/url/
COPY . .
ENV CGO_ENABLED=0
RUN make binaries/linux_x86_64/url

FROM scratch
COPY --from=0 /go/src/github.com/sgreben/url/binaries/linux_x86_64/url /url
ENTRYPOINT [ "/url" ]
