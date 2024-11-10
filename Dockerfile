FROM golang:1.23.2-alpine AS builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

# Use appropriate .env.docker file: if .env.docker does not exist, trying to find .env, if .env does not exist, use .env.default
RUN if [ -f .env.docker ]; then cp .env.docker .env && echo ".env.docker file exists, using .env.docker"; \
	elif [ -f .env ]; then echo ".env.docker file not found, .env file exists, using .env"; \ 
    else cp .env.default .env && echo ".env.docker and .env files not found, using .env.default"; \ 
	fi

RUN go build -o ./bin/bot cmd/bot/main.go


FROM alpine:3.20

WORKDIR /workspace

COPY --from=builder /workspace/bin/bot .
COPY --from=builder /workspace/.env .

CMD [ "./bot" ]