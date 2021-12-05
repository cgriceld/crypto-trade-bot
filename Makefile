PATH = cmd/api

SRC = $(PATH)/main.go $(PATH)/config.go

all:
	go run $(SRC)

test:
	go test ./... -cover

startdb:
	docker-compose up

stopdb:
	docker-compose down

cleandb:
	docker-compose rm -s -f -v

###############

accounts:
	@curl -v 'localhost:5000/accounts'

orders:
	@curl -v 'localhost:5000/orders'

active:
	@curl -v 'localhost:5000/active?market=pi_xbtusd'

activeall:
	@curl -v 'localhost:5000/activeall'

market:
	@curl -v -X POST 'localhost:5000/setmarket?market=pi_xbtusd'

sell:
	@curl -v -X POST 'localhost:5000/setsell?size=1&market=pi_xbtusd&price=50000'

buy:
	@curl -v -X POST 'localhost:5000/setbuy?price=70000&size=1&market=pi_xbtusd'

start:
	@curl -v -X POST 'localhost:5000/start?market=pi_xbtusd'

stop:
	@curl -v -X POST 'localhost:5000/stop?market=pi_xbtusd'

startall:
	@curl -v -X POST 'localhost:5000/startall'

stopall:
	@curl -v -X POST 'localhost:5000/stopall'

unsetsell:
	@curl -v -X POST 'localhost:5000/unsetsell?market=pi_xbtusd'

unsetall:
	@curl -v -X POST 'localhost:5000/unsetall'

running:
	@curl -v 'localhost:5000/running'

###############

.PHONY: all test startdb stopdb cleandb
