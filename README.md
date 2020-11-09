
# Chats

## Environment

KEY              | Description     | Example
-----------------|-----------------|--------
`WEBSOCKET_PORT` | Порт Веб-сокета | `:8002`
`BUS_TOPIC` | Топик шины | `dev_chats.eldar.1.1`
`BUS_TOKEN` | Токен шины |  `gch7t34yur8u4xm37hy7tnh43`
`BUS_URL` | URL шины |  `nats://localhost:4222`
`BUS_TIMEOUT` | Таймут ответа |  `1000`
`DB_DRIVE` | Драйвер БД |  `mysql`
`DB_PORT` | Порт БД |  `3306`
`DB_HOST` | Хост БД |  `localhost`
`DB_USER` | Пользователь БД |  `root`
`DB_PASSWORD` | Пароль |  `root`
`DB_NAME` | Название БД |  `chats`
`DB_PROTOCOL` | Протокол соединения с БД |  `tcp`
`PINBA_HOST` | Хост сервиса мониторинга |  `localhost`
`REDIS_HOST` | Хост KV-БД |  `localhost`
`REDIS_PORT` | Порт KV-БД |  `6379`
`REDIS_TTL` | Время жизни кэша |  `7200`
`SENTRY_PROTOCOL` | Протокол соединения с сервисом логирования |  `http`
`SENTRY_LOGIN` | Логин |  `login`
`SENTRY_PASSWORD` | Пароль |  `password`
`SENTRY_URL` | УРЛ |  `localhost`
`SENTRY_PORT` | Порт |  `9100`
`SENTRY_PROJECT_ID` | Id проекта |  `4`
`RECONNECT_TO_SERVICE` | Шаг переподключения к сервису |  `15`
`SHUTDOWN_SLEEP` | Время на плавную остновку сервиса |  `10`
`SDK_LOG` | Логирование данных через Sdk |  `1`
`CRON` | Включение тикера |  `1`
`CRON_STEP` | Шаг тикера |  `10`

## Bus API

### GET
**`/chats/chats` — Получить список чатов пользователя**
```json
{
  method: "GET",
  path: "/chats/chats",
  body: {
    user_id: uint,
    count: uint16
  }
}
```

**`/chats/chat` — Получить информацию о чате**
```json
{
  method: "GET",
  path: "/chats/chat",
  body: {
    chat_id: uint,
    user_id: uint
  }
}
```

**`/order/chats` — Получить последний чат по заявке для пользователя**
```json
{
  method: "GET",
  path: "/order/chats",
  body: [
    {
      order_id: uint,
      opponent_id: uint
    }...
  ]
}
```

**`/chats/info` — Получить информацию о списке чатов**
```json
{
  method: "GET",
  path: "/chats/info",
  body: {
    user_id: uint,
    chats_id: [
      uint...
    ]
  }
}
```

**`/chats/messages/update` or 
`/chats/chat/recent` — Список сообщений для клиента**
```json
{
  method: "GET",
  path: "/chats/chat/recent",
  body: {
    user_id: uint,
    chat_id: uint,
    message_id: uint,
    count: uint16
  }
}
```
**`/chats/messages/history` or 
`chats/chat/history` — Список истории сообщений для клиента**
```json
{
  method: "GET",
  path: "/chats/chat/history",
  body: {
    user_id: uint,
    chat_id: uint,
    message_id: uint,
    count: uint16
  }
}
```

### POST
**`/chats/new` — Создать чат**
```json
{
  method: "POST",
  path: "/chats/new",
  body: {
    order_id: uint
  }
}
```

**`/chats/user/subscribe` — Подписать пользователя**
```json
{
  method: "POST",
  path: "/chats/user/subscribe",
  body: {
    user_id: uint,
    user_type: string,
    chat_id: uint
  }
}
```

**`/chats/user/unsubscribe` — Отписать пользователя**
```json
{
  method: "POST",
  path: "/chats/user/unsubscribe",
  body: {
    user_id: uint,
    chat_id: uint
  }
}
```
**`/chats/message` — Добавить сообщение**
```json
{
  method: "POST",
  path: "/chats/message",
  body: {
    chat_id: uint,
    user_id: uint,
    message: string,
    client_message_id: string,
    type: string,
    params: [
      string: string...
    ],
    file_id: string
  }
}
```

**`/chats/status` — Изменить статус чата на "закрыт"**
```json
{
  method: "POST",
  path: "/chats/status",
  body: {
    chat_id: uint
  }
}
```
**`/ws/client/message` — Сообщение в веб-сокет для клиента**
```json
{
  method: "POST",
  path: "/ws/client/message",
  body: {
    user_id: uint,
    type: string,
    data: [
      string: string...
    ]
  }
}
```

**`/ws/client/consultation/update` — Изменение статуса заявки**
```json
{
  method: "POST",
  path: "/ws/client/message",
  body: {
    user_id: uint,
    data: {
      active: bool,
      consultationId: uint
    }
  }
}
```

### PUT

**`/chats/user/subscribe` - Изменение userId у подписчика на чат**
```json
{
  method: "PUT",
  path: "/chats/user/subscribe",
  body: {
    chat_id: uint,
    old_user_id: uint,
    new_user_id: uint
  }
}
```

## Http API

### error
***response:***
```json
{
  error: {
    code: int,
    message: string
  }
}
```

### message

***request:***
```json
{
  type: "message",
  data: {
    messages: [
      {
        clientMessageId: string,
        chatId: uint,
        type: string,
        text: string,
        params: [
          string: string...
        ]
      }
    ]
  }
}
```
***response:***
```json
{
  type: "message",
  data: {
    messages: [
      {
        id: uint,
        clientMessageId: string,
        insertDate: string,
        chatId: uint,
        userId: uint,
        sender: string,
        status: string,
        type: string,
        text: string,
        params: [
          string: string...
        ],
        file: {
          id: string,
          title: string,
          url: string,
          thumbnail: string,
          mimeType: string,
          insertDate: string,
          size: int64
        }
      }
    ],
    users: [
      {
        id: uint,
        firstName: string,
        lastName: string,
        middleName: string,
        photo: string
      }
    ]
  }
}
```

### messageStatus
***request:***
```json
{
  type: "messageStatus",
  data: {
    status: string,
    chatId: uint,
    messageId: uint
  }
}
```
***response:***
```json
{
  type: "messageStatus",
  data: {
    status: string,
    chatId: uint,
    messageId: uint
  }
}
```

### opponentStatus
***request:***
```json
{
  type: "opponentStatus",
  data: {
    chatId: uint
  }
}
```
***response:***
```json
{
  type: "opponentStatus",
  data: {
    chatId: uint,
    users: [
      userId: uint,
      status: string
    ]
  }
}
```

### join
***request without response:***
```json
{
  type: "join",
  data: {
    consultationId: uint
  }
}
```

### typing
***request:***
```json
{
  type: "typing",
  data: {
    chatId: uint,
    status: string
  }
}
```
***response:***
```json
{
  type: "typing",
  data: {
    userId: uint,
    message: string,
    status: string
  }
}
```

### consultationUpdate

***response without request:***
```json
{
  type: "consultationUpdate",
  data: {
    active: bool,
    consultationId: uint
  }
}
```

### clientConnectionError
***response without request:***
```json
{
  type: "clientConnectionError",
  data: {
    userId: uint
  }
}