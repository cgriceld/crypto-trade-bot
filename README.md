![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/cgriceld/crypto-trade-bot?logo=Go&style=plastic)

Trade robot on Kraken Futures [demo-platform](https://demo-futures.kraken.com/futures/PI_XBTUSD).

Project tech-features:
* Clean Architecture design pattern
* Graceful Shutdown
* go-chi, Gorilla WebSocket, Postgres (pgx), logrus
* Integration with Telegram Bot
* REST API and Websocket API on Kraken Futures (demo) support
* Unit-tests coverage

❗️ Robot works with API of **demo** version of platform\
❗️ Trading logic of the robot does not guarantee profitability

# contents

1. [Robot Overview](#robot)
2. [Setup](#setup)
3. [Telegram Bot Notifications](#notifications)
4. [Endpoints Documentation](#endpoints)
5. [Launch Example](#launch)

# robot

* The robot uses stop-loss/take-profit strategy. User can configure robot by setting market, price and size. After robot is successfully started, it listens on 1-minute candles (via websocket subscription), compares the average candle price with user settings and sends ioc order on Kraken if the price is triggered.
* The robot can be launched on several markets in parallel.

* For conditions of robot start see /start or /startall endpoint.
* Afer user configured inner robot order (/setsell or /setbuy) this order becomes active. If this order is trigged (sended to Kraken) or explicitly cancelled by the user (/unset...), it becomes inactive. 
* User can set a new order (e.g. set buy order if it was not set at startup), change the price and size in an already active one or cancel the order **when the robot is already running on market** (no need to stop the robot especially for that).

* The robot sends notifications to Telegram bot (see [notifications](#notifications)).
* Information about executed orders is stored in Postgres.

# setup

Parameters must be set as environmental variables (you can use `source setenv.sh` for convenience).

<pre>
APIPublic   - public API-key from Kraken demo-platform
API Private - private API-key from Kraken demo-platform
TgChatID    - Telegram bot chat ID
TgBotURL    - https://api.telegram.org/bot[token]/sendMessage
port        - port on which the robot server will run
dsn         - string for connecting to Postgres
</pre>

Use `docker-compose.yaml` to start Postgres.

Some `Makefile` rules:
* `make`      - start robot server
* `make test` - run tests with coverage

# notifications

`✅ Start subscription on market: pi_ethusd`

The robot is successfuly started on market.

---

`⚠️ Stop subscription on market: pi_ethusd`

The robot on market is stopped.

Reasons:
1. It was explicitly stopped by user (/stop, /stopall).
2. Websocket error (code 1006) and reconnection doesn't help.
3. Server is stopped (graceful shutdown).

---

`📌 Make buy order on pi_xbtusd. Price: 58620.50`

Order was successfuly placed and executed.

---

`❌ Fail to place order: pi_ethusd: sell: server error`

Fail to send order to Kraken due to inner error.

---

`❌ Fail to execute order: pi_xbtusd: sell: insufficient funds`

Order was rejected because of balance error.

---

`❌ Fail to execute order: pi_xbtusd: sell`

Order was rejected because of every another reason except balance error.

#  endpoints

```http
POST /setmarket?market=`market`
```
Sets new market[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 201 (Created)
```

---

```http
POST /setsell?market=`market`&price=`price`&size=`size`
```
Sets inner sell order with passed query parameters. If no orders have been placed on this market before, then you must first set this market (/setmarket)[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "type":"sell", "price":4000, "size":1}, Status 201 (Created)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```
---

```http
POST /setbuy?market=`market`&price=`price`&size=`size`
```
Sets inner buy order with passed query parameters. If no orders have been placed on this market before, then you must first set this market (/setmarket)[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "type":"buy", "price":4000, "size":1}, Status 201 (Created)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```

---

```http
POST /unsetsell?market=`market`
```
Unsets inner sell order on market passed as parameter[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 200 (OK)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```

---

```http
POST /unsetbuy?market=`market`
```
Unsets inner buy order on market passed as parameter[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 200 (OK)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```

---

```http
POST /unsetall
```
Unsets all orders on all previously set markets.

```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"ok"}], Status 200 (OK)
```

---

```http
POST /start?market=`market`
```
Launches the robot on market passed as parameter. The market must be set before (/setmarket), there must be at least one active order on that market and the robot must not be already running on that market[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 200 (OK)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"Fail to start, parameter wasn't set: market"}, Status 400 (Bad Request)
JSON {"market":"pi_ethusd", "status":"Fail to start, parameter wasn't set: pi_ethusd: orders"}, Status 400 (Bad Request)
JSON {"market":"pi_ethusd", "status":"Fail to start, subscription is already running: pi_ethusd"}, Status 400 (Bad Request)
JSON {"market":"pi_ethusd", "status":"Internal Server Error"}, Status 500 (Internal Server Error)
```

---

```http
POST /startall
```
Launches the robot on all previously set markets. The launch criteria are the same as for /start. By the status field in the body of the response, you can understand which markets were launched and which were not and why.
```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"Fail to start, parameter wasn't set: pi_ethusd: orders"}], Status 200 (OK)
```

---

```http
POST /stop?market=`market`
```
Stops the robot on market passed as parameter[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 200 (OK)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```

---

```http
POST /stopall
```
Stops the robot on all previously set markets.

```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"ok"}], Status 200 (OK)
```

---

```http
GET /active?market=`market`
```
Returns currently active orders on market passed as parameter[*](#queries).

```go
Sample Response on Success:
JSON [{"market":"pi_xbtusd", "type":"buy", "price":4000, "size":1}, {"market":"pi_xbtusd", "type":"sell", "price":4000, "size":1}], Status 200 (OK)

Sample Response on Fail:
JSON {"market":"pi_ethusd", "status":"No market was set: pi_ethusd"}, Status 400 (Bad Request)
```

---

```http
GET /activeall
```
Returns all currently active orders on all previously set markets.

```go
JSON [{"market":"pi_xbtusd", "type":"buy", "price":4000, "size":1}, {"market":"pi_ethusd", "type":"buy", "price":4000, "size":1}], Status 200 (OK)
```

---

```http
GET /running
```
Returns markets where the robot is currently running.

```go
JSON [{"market":"pi_xbtusd", "status":"running"}, {"market":"pi_ethusd", "running"}], Status 200 (OK)
```

---

```http
GET /orders
```
Returns executed orders from the database.

```go
Sample Response on Success:
JSON [{"time":"2021-12-01T13:37:37.278283Z", "market":"pi_xbtusd", "type":"buy", "price":56980.5, "size":1}], Status 200 (OK)

Sample Response on Fail:
text/plain Internal Server Error, Status 500 (Internal Server Error)
```

---

```http
GET /accounts
```
Returns user balance info.

```go
Sample Response on Success:
JSON [{"fi_xbtusd":"10"}, {"fi_ethusd":"0"}, ...], Status 200 (OK)

Sample Response on Fail:
text/plain Internal Server Error, Status 500 (Internal Server Error)
```

# queries

In all requests with query parameters the following responses may take place (text/plain):

* `Wrong query parameter: no [market/price/size]`, Status 400 (Bad Request)\
  No parameter
* `Wrong query parameter: [price/size]: [value]`, Status 400 (Bad Request)\
  Invalid parameter value (e.g. negative price)
* `Internal Server Error`, Status 500 (Internal Server Error)\
  Internal error from the middleware during processing
  
# launch

For example, to start robot on pi_ethusd market you should:

1. /setmarket?market=pi_ethusd
2. /setsell?size=5&market=pi_ethusd&price=4000
3. /start?market=pi_ethusd or /startall
4. Profit!
