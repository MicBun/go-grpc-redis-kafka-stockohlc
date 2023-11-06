# go-grpc-redis-kafka-stockohlc
a repository for stock OHLC (Open, High, Low, Close) using grpc, redis, kafka and golang programming language

## Prerequisite
- docker
- docker-compose
- golang

## How to run
The simplest way to run this project is using docker-compose. You can run this project by executing this command:
```
docker-compose up
```
You can also rebuild the image by executing this command:
```
docker-compose up --build
```

## How to test
You can test this project by executing this command:
1. Test server
```
cd server/
sh ./scripts/coverage.sh
```
2. Test client
```
cd client/
sh ./scripts/coverage.sh
```

## Data Flow
Data flow diagram for this project can be seen on file named `data-flow-diagram.png`

## Credits
Made with ❤️ by Michael Buntarman <br>
Visit my LinkedIn at [Michael Buntarman](https://www.linkedin.com/in/MicBun) <br>
Visit my GitHub at [MicBun](https://github.com/MicBun) 
