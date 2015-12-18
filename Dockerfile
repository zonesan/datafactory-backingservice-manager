FROM golang:1.5.1

# for gateway
ENV SERVICE_NAME datafactory_backingserver_manager

ENV SERVICE_PORT 41000
EXPOSE $SERVICE_PORT

ENV SERVICE_SOURCE_URL github.com/asiainfoLDP/datafactory-backingservice-manager

run mkdir $GOPATH/src/$SERVICE_SOURCE_URL -p
WORKDIR $GOPATH/src/$SERVICE_SOURCE_URL

ADD . .

RUN go get github.com/tools/godep &&\
        $GOPATH/bin/godep restore && \
        go build -o backservice-manager

CMD ["sh", "-c", "./backservice-manager"]
