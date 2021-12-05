Trade robot on Kraken Futures [demo-platform](https://demo-futures.kraken.com/futures/PI_XBTUSD). 

# Contents

1. [Robot features](#robot)
2. [Setup](#setup)
3. [Telegram bot notifications](#notifications)
4. [Endpoints documentation](#endpoints)

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

* APIPublic   - public API-key from Kraken demo-platform
* API Private - private API-key from Kraken demo-platform
* TgChatID    - Telegram bot chat ID
* TgBotURL    - URL to the endpoint /SendMessage of Telegram containing the bot token
* port        - the port on which the robot server will run
* dsn         - string for connecting to Postgres

Use `docker-compose.yaml` to start Postgres.

Basic `Makefile` rules:
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
Задает новый рынок[*](#queries).

```go
Sample Response on Success:
JSON {"market":"pi_xbtusd", "status":"ok"}, Status 201 (Created)
```

---

```http
POST /setsell?market=`market`&price=`price`&size=`size`
```
Устанавливает ордер на продажу с переданными параметрами. Если ранее на данном рынке не устанавливалось никаких ордеров, то необходимо сначала установить этот рынок (/setmarket)[*](#queries).

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
Устанавливает ордер на покупку с переданными параметрами. Если ранее на данном рынке не устанавливалось никаких ордеров, то необходимо сначала установить этот рынок (/setmarket)[*](#queries).

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
Отменяет ордер на продажу на данном рынке[*](#queries).

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
Отменяет ордер на покупку на данном рынке[*](#queries).

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
Отменяет все ордеры на всех рынках.

```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"ok"}], Status 200 (OK)
```

---

```http
POST /start?market=`market`
```
Запускает робота на данном рынке. Для запуска необходимо, чтобы рынок был задан ранее (/setmarket), на нем был задан хотя бы один ордер (на покупку или продажу), на нем уже не был запущен робот[*](#queries).

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
Запускает робота на всех рынках. Критерии запуска такие же, как и при /start. По статусу в теле ответа можно понять, какие рынки были запущены, а какие - нет и почему.

```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"Fail to start, parameter wasn't set: pi_ethusd: orders"}], Status 200 (OK)
```

---

```http
POST /stop?market=`market`
```
Останавливает робота на данном рынке[*](#queries).

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
Останавливает робота на всех рынках.

```go
JSON [{"market":"pi_xbtusd", "status":"ok"}, {"market":"pi_ethusd", "status":"ok"}], Status 200 (OK)
```

---

```http
GET /active?market=`market`
```
Возвращает активные на данный момент ордеры на данном рынке[*](#queries).

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
Возвращает активные на данный момент ордеры на всех рынках.

```go
JSON [{"market":"pi_xbtusd", "type":"buy", "price":4000, "size":1}, {"market":"pi_ethusd", "type":"buy", "price":4000, "size":1}], Status 200 (OK)
```

---

```http
GET /running
```
Возвращает, на каких рынках в данный момент запущен робот.

```go
JSON [{"market":"pi_xbtusd", "status":"running"}, {"market":"pi_ethusd", "running"}], Status 200 (OK)
```

---

```http
GET /orders
```
Возвращает выставленные и исполненные ордеры.

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
Возвращает информацию о балансе пользователя.

```go
Sample Response on Success:
JSON [{"fi_xbtusd":"10"}, {"fi_ethusd":"0"}, ...], Status 200 (OK)

Sample Response on Fail:
text/plain Internal Server Error, Status 500 (Internal Server Error)
```

# queries

Во всех запросах, где передаются query-параметры, в случае ошибки обработки параметров возможны следующие response (text/plain).

* `Wrong query parameter: no [market/price/size]`, Status 400 (Bad Request)\
  Отсутствует параметр
* `Wrong query parameter: [price/size]: [value]`, Status 400 (Bad Request)\
  Недопустимое значение параметра (отрицательное число, 0)
* `Internal Server Error`, Status 500 (Internal Server Error)\
  Внутренняя ошибка при обработке параметров
