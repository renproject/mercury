FROM ubuntu:xenial

# Install zcashd
RUN apt-get update && \
    apt-get install -y --no-install-recommends apt-transport-https gnupg2 ca-certificates wget && \
    wget -qO - https://apt.z.cash/zcash.asc | apt-key add - && \
    echo "deb https://apt.z.cash/ jessie main" | tee /etc/apt/sources.list.d/zcash.list && \
    apt-get update && apt-get install -y --no-install-recommends zcash && \
    mkdir -p /root/.zcash-params && zcash-fetch-params

ARG Address

COPY zcash.conf /root/.zcash/
RUN echo "mineraddress=${Address}" >> ~/.zcash/zcash.conf
COPY zcash.sh /root/
RUN echo "#!/bin/bash\nAddress=${Address}"|cat - ~/zcash.sh > /tmp/out && mv /tmp/out ~/zcash.sh
RUN chmod +x ~/zcash.sh

EXPOSE 18232

CMD ["./root/zcash.sh"]