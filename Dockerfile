FROM alpine
ARG BIN_NAME=node
WORKDIR /app
COPY /artifacts/$BIN_NAME /app/node
EXPOSE 8015
ENTRYPOINT ["./node"]
