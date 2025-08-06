# Webhook Telegram Bot - Production Deployment

Автоматическая система для отправки уведомлений о новых постах с форума в Telegram.

## 🚀 Быстрая установка на сервере

### 1. Скачать и установить

```bash
# Скачать установщик
curl -fsSL https://raw.githubusercontent.com/dignezzz/webhook_tg_bot/main/install.sh -o install.sh
chmod +x install.sh

# Запустить установку
./install.sh install
```

### 2. Настроить конфигурацию

Скрипт автоматически создаст файл `.env` и попросит заполнить параметры:

- **TELEGRAM_BOT_TOKEN** - токен Telegram бота
- **TELEGRAM_CHAT_ID** - ID чата для уведомлений  
- **WEBHOOK_SECRET** - секретный ключ для webhook'ов
- **OPENAI_API_KEY** - ключ OpenAI для генерации резюме
- **BASE_URL** - адрес вашего форума

### 3. Настроить webhook на форуме

В админке Discourse добавьте webhook:
- **URL**: `http://your-server.com:8080/webhook`
- **Secret**: тот же что в WEBHOOK_SECRET
- **Events**: Topic Created, Post Created

## 🛠️ Управление

### Интерактивное меню
```bash
whtg
```

### Команды
```bash
whtg install    # Установить бота
whtg update     # Обновить до последней версии
whtg start      # Запустить бота
whtg stop       # Остановить бота
whtg restart    # Перезапустить бота
whtg status     # Показать статус
whtg logs       # Показать логи
whtg uninstall  # Удалить бота
```

## 🔧 Дополнительные настройки

### Игнорирование пользователей
```bash
# В .env файле
IGNORED_USERS=-2,-1  # ID ботов (discobot, chatbot)
```

### Категории тем Telegram
```bash
# Пример направления разных категорий в разные топики
TELEGRAM_THREAD_ID_1=123456
THREAD_CATEGORIES_1=1,2,3

TELEGRAM_THREAD_ID_2=234567
THREAD_CATEGORIES_2=4,5
```

### Платные разделы
```bash
PREMIUM_CATEGORIES=4,5,6  # ID категорий с подпиской
```

## 📋 Файлы и директории

- **Установка**: `/opt/webhook_tg_bot/`
- **Конфигурация**: `/opt/webhook_tg_bot/.env`
- **Docker Compose**: `/opt/webhook_tg_bot/docker-compose.yml`
- **Systemd сервис**: `/etc/systemd/system/webhook-tg-bot.service`
- **Docker образ**: `ghcr.io/dignezzz/webhook_tg_bot:latest`

## 🐳 Только Docker (без установщика)

```bash
# Создать docker-compose.yml
version: '3.8'
services:
  webhook-bot:
    image: ghcr.io/dignezzz/webhook_tg_bot:latest
    container_name: webhook_tg_bot
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env

# Запустить
docker-compose up -d
```

## 🔄 Автоматические обновления

GitHub Actions автоматически собирает новые образы при push в main ветку.
Для обновления используйте: `whtg update`

## 📝 Логи и мониторинг

```bash
# Посмотреть логи
whtg logs

# Проверить состояние
whtg status

# Healthcheck endpoint
curl http://localhost:8080/health
```

## 🆘 Поддержка

- **GitHub**: https://github.com/dignezzz/webhook_tg_bot
- **Issues**: https://github.com/dignezzz/webhook_tg_bot/issues
