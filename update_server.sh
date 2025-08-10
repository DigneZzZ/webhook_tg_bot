#!/bin/bash

# Скрипт для обновления webhook_tg_bot на сервере

echo "🔄 Updating webhook_tg_bot with official OpenAI SDK..."

# Остановка контейнера
echo "📦 Stopping container..."
docker compose down

# Обновление кода (если используется git)
echo "🔄 Pulling latest code..."
git pull origin main

# Пересборка образа
echo "🏗️ Rebuilding Docker image..."
docker compose build --no-cache

# Запуск контейнера
echo "🚀 Starting container..."
docker compose up -d

# Показ логов
echo "📋 Showing logs..."
docker compose logs -f --tail=50

echo "✅ Update completed!"
