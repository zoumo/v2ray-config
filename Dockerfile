FROM v2ray/official:latest

RUN apk --no-cache add bash

COPY output/* /etc/v2ray/

ENTRYPOINT [ "v2ray" ]

CMD ["-config=/etc/v2ray/config-server.json"]
