FROM ubuntu:xenial

# Install bitcoind from PPA
RUN apt-get update && apt-get install --yes software-properties-common && \
    add-apt-repository --yes ppa:bitcoin/bitcoin && apt-get update && \
    apt-get install --yes bitcoind

COPY bitcoin.conf /root/.bitcoin/
COPY run.sh /root/
RUN chmod +x ~/run.sh
EXPOSE 18443

ENTRYPOINT ["./root/run.sh", "mmjwW62DSSWbHuMmRTqUdWd5ABGStniygb"]