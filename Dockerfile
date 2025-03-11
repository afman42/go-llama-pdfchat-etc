FROM ubuntu:23.10

COPY ./pdf2txt /usr/local/bin/

WORKDIR /app

COPY ./bin/linux_amd64/app ./
COPY ./web/dist ./web/dist
COPY .env.prod .env.prod

EXPOSE 9002

CMD ["./app","-mode","prod"]
