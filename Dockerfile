FROM alpine

WORKDIR /app

COPY build/url_shortener_linux /app/url_shortener

EXPOSE 8080

CMD ["/app/url_shortener"]
