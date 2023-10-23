FROM golang:alpine

WORKDIR /root
COPY ./ /root/
RUN go mod download && go mod verify && go build .

ENTRYPOINT /root/synok
