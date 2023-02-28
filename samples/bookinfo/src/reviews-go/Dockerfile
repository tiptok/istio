FROM golang:1.18 as builder
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn
WORKDIR /opt/app
COPY . .
RUN CGO_ENABLED=0 go build -o reviews

FROM ubuntu:latest
WORKDIR /opt/app

ARG service_version
ARG enable_ratings
ARG star_color
ENV SERVICE_VERSION ${service_version:-v1}
ENV ENABLE_RATINGS ${enable_ratings:-false}
ENV STAR_COLOR ${star_color:-black}

COPY --from=builder /opt/app/reviews ./reviews
CMD ["/opt/app/reviews"]