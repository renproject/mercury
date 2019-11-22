FROM node:alpine
WORKDIR /app
# Install ganache-cli globally
RUN npm install -g ganache-cli
EXPOSE 8545
CMD ["ganache-cli", "--account=0xDA42F3D39085CBCA8B145B92BB4E2F226CB5DC78934D31E6BF3F5DD59D1D7553,1000000000000000000000", "--gasPrice=0", "--h=0.0.0.0"]