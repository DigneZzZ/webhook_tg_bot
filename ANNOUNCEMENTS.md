# Автоматические анонсы премиум контента

## Описание функциональности

Теперь бот может автоматически создавать анонсы тем из платных категорий в отдельной категории форума. Это позволяет привлекать внимание к премиум контенту и стимулировать подписки.

## Как это работает

1. **Обнаружение премиум темы**: Когда создается новая тема в категории, указанной в `PREMIUM_CATEGORIES`
2. **Создание анонса**: Автоматически создается анонс в категории, указанной в `ANNOUNCEMENT_CATEGORY_ID`
3. **Форматирование**: Анонс содержит краткое описание, информацию об авторе и призыв к подписке

## Новые настройки в .env

```bash
# Discourse API Configuration (для создания анонсов)
DISCOURSE_API_KEY=your_discourse_api_key_here
DISCOURSE_API_USERNAME=your_discourse_username_here  
DISCOURSE_BASE_URL=https://your-forum.com

# Настройки анонсов
ENABLE_ANNOUNCEMENTS=true                # Включить автоматические анонсы
ANNOUNCEMENT_CATEGORY_ID=15              # ID категории для размещения анонсов
```

### Получение API ключа Discourse

1. Перейдите в админку Discourse: `https://your-forum.com/admin/api/keys`
2. Нажмите "New API Key"
3. Заполните параметры:
   - **Description**: `webhook_tg_bot_announcements`
   - **User Level**: Используйте аккаунт администратора или специального бота
   - **Scopes**: 
     - ✅ `topics:write` - для создания тем
     - ✅ `categories:read` - для получения списка категорий
4. Скопируйте сгенерированный ключ в `DISCOURSE_API_KEY`
5. Укажите username пользователя в `DISCOURSE_API_USERNAME`

### Определение ID категории для анонсов

1. Создайте категорию "Анонсы премиум контента" в Discourse
2. Получите её ID одним из способов:
   - URL категории: `https://your-forum.com/c/announcements/15` (где 15 - это ID)
   - Админка: `https://your-forum.com/admin/customize/site_texts`
   - API: `curl -H "Api-Key: YOUR_KEY" -H "Api-Username: USERNAME" https://your-forum.com/categories.json`

## Пример анонса

```markdown
## 👑 AdminUser создал новую тему в премиум разделе

**Тема:** Продвинутые техники программирования на Go

**Краткое описание:**
В этой теме мы рассмотрим сложные паттерны проектирования и оптимизацию производительности...

**Теги:** #golang, #advanced, #patterns

**Для доступа к полной теме** и другим материалам премиум раздела, оформите VIP подписку в нашем боте: @gig_combot

---

💎 **Это анонс платного контента.** Полная тема доступна только подписчикам VIP.

🔗 **[Перейти к теме](https://forum.example.com/t/topic/123)** (требуется подписка)
```

## Логика работы

### Условия создания анонса:
- ✅ `ENABLE_ANNOUNCEMENTS=true`
- ✅ Настроен `ANNOUNCEMENT_CATEGORY_ID`
- ✅ Тема создана в категории из `PREMIUM_CATEGORIES`
- ✅ Настроены API ключи Discourse
- ✅ Тема НЕ создается в самой категории анонсов

### Безопасность:
- Используется отдельный API ключ с ограниченными правами
- Проверка подключения к Discourse API при запуске
- Логирование всех операций создания анонсов
- Graceful degradation - если анонс не удается создать, основная функциональность работает

## Установка и настройка

### 1. Обновите конфигурацию
```bash
# Скопируйте новый пример
cp .env.example .env.new
# Перенесите свои настройки и добавьте новые
```

### 2. Настройте Discourse API
- Создайте API ключ в Discourse
- Создайте категорию для анонсов
- Обновите .env файл

### 3. Перезапустите бота
```bash
# Если используете Docker Compose
docker-compose down
docker-compose up -d --build

# Если используете systemd
sudo systemctl restart webhook-tg-bot
```

### 4. Проверьте логи
```bash
# Docker
docker-compose logs -f webhook-bot

# Systemd  
sudo journalctl -u webhook-tg-bot -f
```

## Отладка

### Проверка подключения к Discourse API
```bash
curl -H "Api-Key: YOUR_API_KEY" \
     -H "Api-Username: YOUR_USERNAME" \
     https://your-forum.com/categories.json
```

### Тестирование создания анонса
1. Создайте тему в премиум категории
2. Проверьте логи бота на наличие сообщений об анонсах
3. Убедитесь, что анонс появился в категории анонсов

### Частые ошибки
- **403 Forbidden**: Неверный API ключ или недостаточно прав
- **404 Not Found**: Неверный URL категории или она не существует
- **422 Unprocessable Entity**: Ошибка в данных запроса (обычно проблема с форматированием)

## Мониторинг

В логах будут отображаться:
- ✅ Успешное создание анонсов: `Created announcement topic: [title] (ID: [id])`
- ⚠️ Предупреждения: `Warning: Failed to initialize announcement service`
- ❌ Ошибки: `Failed to create announcement for topic [id]: [error]`

## Отключение анонсов

Чтобы отключить автоматические анонсы:
```bash
ENABLE_ANNOUNCEMENTS=false
```

Или удалите/закомментируйте настройки:
```bash
# DISCOURSE_API_KEY=...
# DISCOURSE_API_USERNAME=...
# ANNOUNCEMENT_CATEGORY_ID=...
```