FROM ubuntu:xenial

# Install zcashd
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl

RUN bash <(curl https://get.parity.io -L)

ARG Address

COPY zcash.conf /root/.zcash/
COPY mine.sh /root/
RUN chmod +x ~/mine.sh
EXPOSE 18443

CMD ["./root/mine.sh", "${Address}"]