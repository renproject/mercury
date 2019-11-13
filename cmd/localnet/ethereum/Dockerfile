FROM ubuntu:xenial

RUN apt-get update && apt-get install -y --no-install-recommends wget ca-certificates
RUN wget https://releases.parity.io/ethereum/v2.6.4/x86_64-unknown-linux-gnu/parity && chmod +x parity && cp parity /usr/local/bin

EXPOSE 8545

CMD ["parity", "--chain", "dev", "--jsonrpc-cors", "all", "--jsonrpc-interface", "all", "--jsonrpc-hosts", "all"]