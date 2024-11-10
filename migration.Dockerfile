FROM alpine:3.20

RUN apk update && \
    apk upgrade && \
    apk add bash && \
    rm -rf /var/cache/apk/*

ADD https://github.com/pressly/goose/releases/download/v3.20.0/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

WORKDIR /workspace

ADD migration/*.sql migration/
ADD migration.sh .
ADD .env* .

# Use appropriate .env.docker file: if .env.docker does not exist, trying to find .env, if .env does not exist, use .env.default
RUN if [ -f .env.docker ]; then cp .env.docker .env && echo ".env.docker file exists, using .env.docker"; \
	elif [ -f .env ]; then echo ".env.docker file not found, .env file exists, using .env"; \ 
    else cp .env.default .env && echo ".env.docker and .env files not found, using .env.default"; \ 
	fi

RUN chmod +x migration.sh

ENTRYPOINT ["bash", "migration.sh"]