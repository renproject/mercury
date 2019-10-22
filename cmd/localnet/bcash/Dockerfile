FROM ubuntu:xenial

# Install bitcoind from PPA
RUN apt-get update && apt-get install --yes software-properties-common && \
    add-apt-repository --yes ppa:bitcoin-abc/ppa && apt-get update && \
    apt-get install --yes bitcoind

ARG Address

COPY bitcoin.conf /root/.bitcoin/
COPY mine.sh /root/
RUN echo "#!/bin/bash\nAddress=${Address}"|cat - ~/mine.sh > /tmp/out && mv /tmp/out ~/mine.sh
RUN chmod +x ~/mine.sh

EXPOSE 18443

CMD ["./root/mine.sh"]