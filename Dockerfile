# develop
FROM golang:1.12-alpine as build
WORKDIR /go/linebot-sample
COPY . .
RUN apk add --no-cache git make && go get github.com/oxequa/realize && make build

# ecr
FROM alpine
WORKDIR /linebot-sample
COPY --from=build /go/linebot-sample/bin/api .
COPY --from=build /go/linebot-sample/_tools ./_tools
RUN addgroup api && adduser -D -G api api && chown -R api:api /linebot-sample/api
CMD ["./api"]
