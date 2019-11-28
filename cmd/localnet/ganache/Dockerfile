FROM mhart/alpine-node:11.10.1

RUN mkdir -p /home/root/.ganache

# Install ganache-cli globally
RUN yarn global add ganache-cli@6.5.1
EXPOSE 8545

ENTRYPOINT [ "ganache-cli", "-i", "1313", "-d" ,"-p" ,"8545" ,"-h" ,"0.0.0.0" ,"--db=/home/root/.ganache"]