FROM alpine:latest

RUN mkdir /app

COPY restApp /app

CMD [ "/app/restApp"]