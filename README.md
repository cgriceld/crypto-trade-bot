# setup

Перед началом работы необходимо задать ряд параметров через переменные окружения, для этого можно воспользоваться файлом `setenv.sh` (в корне).
Параметры:
* APIPublic - публичный ключ API от Kraken
* APIPrivate - приватный ключ API от Kraken
* TgChatID - ID чата Telegram-бота
* TgBotURL - URL на endpoint /sendMessage у Telegram, содержащий токен бота
* port - порт, на котором будет запущен сервер
* dsn - строка для подключения к Postgres

Postgres запускается через `docker-compose.yaml` (в корне).

В корне репозитория также есть `Makefile`:
* `make` - запускает сервер
* `make test` - запускает тесты с coverage
* `make startdb` - поднимает базу
* `make stopdb` - останавливает базу

#  robot

* Робот реализует стратегию stop-loss/take-profit. Пользователь выбирает рынок, цену, при которой он бы хотел продать или купить, и размер (сколько продать/купить). Робот слушает одноминутные свечи, вычисляет среднюю цену по свече, сравнивает ее с установленными пользователем значениями и отправляет ордер на Kraken, если триггер срабатывает.

* Робот может быть одновременно запущен на нескольких рынках.

* Для запуска робота на рынке необходимо, чтобы рынок был задан ранее (/setmarket), на нем был задан хотя бы один ордер (/setsell, /sellbuy), на нем уже не был запущен робот (подробнее см. endpoints /start или /startall).

* После настройки ордера (/setsell или /setbuy) ордер становится активным. Когда он срабатывает (отправляется на Kraken) или отменяется самим пользователем (/unsetsell, /unsetbuy, /unsetall), он становится неактивным. Задать новый ордер, поменять цену и размер в уже активном, отменить ордер **можно при запущенном роботе** (для этого не надо его специально останаливать).

Например, для запуска робота на рынке pi_ethusd необходимо выполнить следующие действия:
1. /setmarket?market=pi_ethusd
2. /setsell?size=5&market=pi_ethusd&price=4000
3. /start?market=pi_ethusd или /startall
4. Profit!

# notifications

Пользователю отправляется несколько видом уведомлений.

`✅ Start subscription on market: pi_ethusd`

Робот запустился на рынке.

---

`⚠️ Stop subscription on market: pi_ethusd`

Робот на рынке остановлен.
Это может быть причине:
1. Явной остановки (/stop, /stopall)
2. Произошла ошибка веб-сокета и переподключиться не получилось
3. Сервер остановлен (sigint)

---

`📌 Make buy order on pi_xbtusd. Price: 58620.50`

Ордер успешно исполнен.

---

`❌ Fail to place order: pi_ethusd: sell: server error`

Не удалось отправить ордер на Kraken по причине внутренней ошибки.

---

`❌ Fail to execute order: pi_xbtusd: sell: insufficient funds`

Kraken не может исполнить ордер, так как на балансе недостаточно средств.

---

`❌ Fail to execute order: pi_xbtusd: sell`

Kraken не может исполнить ордер по любой другой причине, отличной от недостаточного баланса.

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