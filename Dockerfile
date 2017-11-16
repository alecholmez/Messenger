FROM iron/go:dev

WORKDIR /gocode/src/github.com/alecholmez/messenger

COPY . /gocode/src/github.com/alecholmez/messenger
RUN go build -o /usr/local/bin/messenger

CMD ["messenger"]
EXPOSE 8080
