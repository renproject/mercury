FROM ubuntu:xenial

COPY mercury /root/
RUN chmod +x ~/mercury

ENV BTC_RPC_URL="http://bitcoin:18443"
ENV ZEC_RPC_URL="http://zcash:18232"
ENV BCH_RPC_URL="http://bcash:18443"
ENV ETH_RPC_URL="http://ethereum:8545"

EXPOSE 5000

CMD ["./root/mercury"]