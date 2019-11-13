FROM ubuntu:xenial

# Install zcashd
RUN apt-get update && \
    apt-get install -y --no-install-recommends apt-transport-https gnupg2 ca-certificates wget && \
    wget -qO - https://apt.z.cash/zcash.asc | apt-key add - && \
    echo "deb https://apt.z.cash/ jessie main" | tee /etc/apt/sources.list.d/zcash.list && \
    apt-get update && apt-get install -y --no-install-recommends zcash && \
    mkdir -p /root/.zcash-params && zcash-fetch-params

COPY zcash.conf /root/.zcash/
COPY run.sh /root/
RUN chmod +x ~/run.sh

EXPOSE 18232

ENTRYPOINT ["./root/run.sh", "tmExAcmdSvXsNodpDRbg6LAMUqeVyMVtikL"]