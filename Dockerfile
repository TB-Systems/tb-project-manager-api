ARG GO_VERSION=1.25.5

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/tb-project-manager-api .

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /bin/tb-project-manager-api /app/tb-project-manager-api

ENV APP_ENV=production
ENV PORT=3000

EXPOSE 3000

USER app

CMD ["/app/tb-project-manager-api"]
