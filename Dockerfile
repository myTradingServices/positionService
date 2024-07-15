FROM golang:1.21

WORKDIR /go/project/positionService

ADD go.mod go.sum main.go ./
ADD internal ./internal
ADD proto ./proto
ADD migrations ./migrations


EXPOSE 7071 5432 7074 7075

CMD ["go", "run", "main.go"]
