# Быстрая установка на Ubuntu

## Автоматическая установка

```bash
# Скачиваем и запускаем скрипт установки
wget https://raw.githubusercontent.com/your-repo/webhook_tg_bot/main/deploy_ubuntu.sh
chmod +x deploy_ubuntu.sh
./deploy_ubuntu.sh
```

## Ручная установка

### 1. Установка зависимостей

```bash
# Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Перелогиньтесь для применения прав docker
```

### 2. Подготовка

```bash
# Создаем директорию
sudo mkdir -p /opt/webhook_tg_bot
sudo chown $USER:$USER /opt/webhook_tg_bot
cd /opt/webhook_tg_bot

# Клонируем репозиторий (или загружаем файлы)
git clone https://github.com/your-repo/webhook_tg_bot.git .
```

### 3. Конфигурация

```bash
# Копируем пример конфигурации
cp .env.example .env

# Редактируем конфигурацию
nano .env
```

Заполните следующие параметры:
```env
TELEGRAM_BOT_TOKEN=123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZ
TELEGRAM_CHAT_ID=-1001234567890
TELEGRAM_THREAD_ID=0
WEBHOOK_SECRET=your_super_secret_key_here
WEBHOOK_PORT=8080
OPENAI_API_KEY=sk-your_openai_api_key_here
OPENAI_MODEL=gpt-4o-mini
BASE_URL=https://your-forum.com
PREMIUM_CATEGORIES=4,5,6
CATEGORY_MAPPING={"1":"Общие вопросы","2":"Техподдержка","3":"Личные"}
```

### 4. Запуск

```bash
# Сборка и запуск
docker-compose up -d --build

# Проверка статуса
docker-compose ps
docker-compose logs -f
```

### 5. Автозапуск

```bash
# Создаем systemd сервис
sudo tee /etc/systemd/system/webhook-tg-bot.service > /dev/null << EOF
[Unit]
Description=Webhook Telegram Bot
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=true
WorkingDirectory=/opt/webhook_tg_bot
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Включаем автозапуск
sudo systemctl enable webhook-tg-bot.service
sudo systemctl start webhook-tg-bot.service
```

## Управление ботом

```bash
# Просмотр логов
docker-compose logs -f

# Перезапуск
docker-compose restart

# Остановка
docker-compose down

# Статус
docker-compose ps

# Обновление
git pull
docker-compose up -d --build
```

## Настройка Discourse

1. Перейдите в **Admin → API → Webhooks**
2. Создайте новый webhook:
   - **Payload URL**: `http://your-server-ip:8080/webhook`
   - **Content Type**: `application/json`
   - **Secret**: значение из `WEBHOOK_SECRET`
   - **Events**: выберите `Topic Event` и `Post Event`

## Проверка работы

```bash
# Проверка health endpoint
curl http://localhost:8080/health

# Тест webhook
curl -X POST "http://localhost:8080/webhook?secret=your_secret" \
  -H "Content-Type: application/json" \
  -d '{"topic":{"id":1,"title":"Test","created_by":{"username":"test"},"category_id":1,"tags":["test"]}}'
```

## Решение проблем

### Бот не отвечает
```bash
# Проверяем логи
docker-compose logs webhook-bot

# Проверяем конфигурацию
cat .env

# Перезапускаем
docker-compose restart
```

### Ошибки сборки
```bash
# Очищаем кеш Docker
docker system prune -a
docker-compose build --no-cache
```

### Проблемы с правами
```bash
# Проверяем права на директорию
ls -la /opt/webhook_tg_bot

# Исправляем права
sudo chown -R $USER:$USER /opt/webhook_tg_bot
```
