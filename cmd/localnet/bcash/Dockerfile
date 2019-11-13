FROM ubuntu:xenial

# Install bitcoind from PPA
RUN apt-get update && apt-get install --yes software-properties-common && \
    add-apt-repository --yes ppa:bitcoin-abc/ppa && apt-get update && \
    apt-get install --yes bitcoind

COPY bitcoin.conf /root/.bitcoin/
COPY run.sh /root/
RUN chmod +x ~/run.sh
EXPOSE 18443

ENTRYPOINT ["./root/run.sh", "qzjwrklk7mr5qnhp66z7vy50gj0tnj3xxgl5ztpnsx"]