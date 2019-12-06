FROM golang:alpine

ADD . /home
        
WORKDIR /home

RUN \
       apk add --no-cache bash git openssh && \
       go get -u github.com/prometheus/client_golang/prometheus 
       
CMD ["go","run","main.go"]