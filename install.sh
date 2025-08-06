#!/bin/bash

# Webhook Telegram Bot Installer & Manager
# Version: 2.0 - Production Ready

SCRIPT_NAME="whtg"
INSTALL_PATH="/usr/local/bin/$SCRIPT_NAME"
SERVICE_NAME="webhook-tg-bot"
BOT_DIR="/opt/webhook_tg_bot"
GITHUB_REPO="dignezzz/webhook_tg_bot"
DOCKER_IMAGE="ghcr.io/dignezzz/webhook_tg_bot:latest"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Функции для вывода
print_header() {
    echo -e "${BOLD}${BLUE}=================================================="
    echo -e "    Webhook Telegram Bot Manager v2.0"
    echo -e "==================================================${NC}"
}

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}🎉${NC} $1"
}

# Функция для генерации секретного ключа
generate_secret() {
    if command -v openssl &> /dev/null; then
        openssl rand -hex 32
    elif [ -r /dev/urandom ]; then
        dd if=/dev/urandom bs=32 count=1 2>/dev/null | xxd -p -c 32
    else
        echo "wh_$(date +%s)_$(whoami)_$(hostname)_$(( RANDOM * RANDOM ))" | sha256sum | cut -d' ' -f1
    fi
}

# Проверка Docker Compose команды
check_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif docker compose version &> /dev/null; then
        echo "docker compose"
    else
        echo ""
    fi
}

# Получение статуса сервиса
get_service_status() {
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo -e "${GREEN}Запущен${NC}"
    elif systemctl is-enabled --quiet $SERVICE_NAME; then
        echo -e "${YELLOW}Остановлен (автозапуск включен)${NC}"
    else
        echo -e "${RED}Не установлен${NC}"
    fi
}

# Получение статуса контейнера
get_container_status() {
    if [ -d "$BOT_DIR" ]; then
        cd "$BOT_DIR"
        COMPOSE_CMD=$(check_compose_cmd)
        if [ -n "$COMPOSE_CMD" ]; then
            if $COMPOSE_CMD ps 2>/dev/null | grep -q "Up"; then
                echo -e "${GREEN}Работает${NC}"
            else
                echo -e "${RED}Остановлен${NC}"
            fi
        else
            echo -e "${RED}Docker Compose недоступен${NC}"
        fi
    else
        echo -e "${RED}Не установлен${NC}"
    fi
}

# Показать статус системы
show_status() {
    print_header
    echo
    echo -e "${BOLD}📊 Текущий статус системы:${NC}"
    echo
    echo -e "   Сервис systemd:     $(get_service_status)"
    echo -e "   Контейнер Docker:   $(get_container_status)"
    
    if [ -f "$BOT_DIR/.env" ]; then
        WEBHOOK_PORT=$(grep "^WEBHOOK_PORT=" "$BOT_DIR/.env" | cut -d'=' -f2)
        WEBHOOK_PATH=$(grep "^WEBHOOK_PATH=" "$BOT_DIR/.env" | cut -d'=' -f2)
        BASE_URL=$(grep "^BASE_URL=" "$BOT_DIR/.env" | cut -d'=' -f2)
        echo -e "   Порт webhook:       ${CYAN}$WEBHOOK_PORT${NC}"
        echo -e "   Путь webhook:       ${CYAN}$WEBHOOK_PATH${NC}"
        echo -e "   Форум:              ${CYAN}$BASE_URL${NC}"
    fi
    
    if [ -d "$BOT_DIR" ]; then
        echo -e "   Директория:         ${CYAN}$BOT_DIR${NC}"
        echo -e "   Конфигурация:       ${CYAN}$BOT_DIR/.env${NC}"
    fi
    echo
}

# Показать интерактивное меню
show_menu() {
    show_status
    echo -e "${BOLD}🛠️  Управление ботом:${NC}"
    echo
    echo -e "   ${CYAN}1.${NC} Установить бота"
    echo -e "   ${CYAN}2.${NC} Обновить бота"
    echo -e "   ${CYAN}3.${NC} Запустить бота"
    echo -e "   ${CYAN}4.${NC} Остановить бота"
    echo -e "   ${CYAN}5.${NC} Перезапустить бота"
    echo -e "   ${CYAN}6.${NC} Показать логи"
    echo -e "   ${CYAN}7.${NC} Обновить статус"
    echo -e "   ${CYAN}8.${NC} Редактировать конфигурацию"
    echo -e "   ${CYAN}9.${NC} Удалить бота"
    echo -e "   ${CYAN}0.${NC} Выход"
    echo
    echo -ne "${BOLD}Выберите действие [0-9]: ${NC}"
}

# Установка бота
install_bot() {
    print_header
    echo
    print_info "Начинаем установку Webhook Telegram Bot..."
    echo

    # Проверка зависимостей
    print_info "1. Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker не установлен"
        echo "Установите Docker:"
        echo "curl -fsSL https://get.docker.com | sh"
        echo "sudo usermod -aG docker \$USER"
        return 1
    fi
    print_status "Docker установлен"
    
    COMPOSE_CMD=$(check_compose_cmd)
    if [ -z "$COMPOSE_CMD" ]; then
        print_error "Docker Compose не установлен"
        echo "Установите Docker с Compose:"
        echo "curl -fsSL https://get.docker.com | sh"
        echo "sudo usermod -aG docker \$USER"
        return 1
    fi
    print_status "Docker Compose установлен"

    # Создание рабочей директории
    print_info "2. Создание рабочей директории..."
    ORIGINAL_PWD="$(pwd)"
    
    if [ ! -d "$BOT_DIR" ]; then
        if [[ $EUID -eq 0 ]]; then
            mkdir -p "$BOT_DIR"
        else
            sudo mkdir -p "$BOT_DIR"
            sudo chown $USER:$USER "$BOT_DIR"
        fi
        print_status "Создана директория $BOT_DIR"
    else
        print_warning "Директория $BOT_DIR уже существует"
    fi
    
    cd "$BOT_DIR"

    # Конфигурация
    print_info "3. Настройка конфигурации..."
    
    if [ ! -f ".env" ]; then
        echo
        echo "Настройка конфигурации бота:"
        echo
        
        # Telegram настройки
        read -p "Введите Telegram Bot Token: " TELEGRAM_BOT_TOKEN
        read -p "Введите Telegram Chat ID: " TELEGRAM_CHAT_ID
        read -p "Введите Telegram Thread ID (0 если не нужен): " TELEGRAM_THREAD_ID
        
        echo
        # Webhook настройки
        WEBHOOK_SECRET=$(generate_secret)
        echo "Сгенерирован секретный ключ для webhook: ${WEBHOOK_SECRET:0:16}..."
        
        WEBHOOK_PATH="/$(openssl rand -hex 5 2>/dev/null || echo "$(date +%s)$(( RANDOM ))" | sha256sum | cut -c1-10)"
        echo "Сгенерирован путь для webhook: $WEBHOOK_PATH"
        
        read -p "Введите порт для webhook (по умолчанию 8080): " WEBHOOK_PORT
        WEBHOOK_PORT=${WEBHOOK_PORT:-8080}
        
        echo
        # AI настройки
        read -p "Введите OpenAI API Key: " OPENAI_API_KEY
        read -p "Введите OpenAI модель (по умолчанию gpt-4o-mini): " OPENAI_MODEL
        OPENAI_MODEL=${OPENAI_MODEL:-gpt-4o-mini}
        
        echo
        # Форум настройки
        read -p "Введите базовый URL вашего форума: " BASE_URL
        read -p "Введите ID платных категорий (Enter для пропуска): " PREMIUM_CATEGORIES
        PREMIUM_CATEGORIES=${PREMIUM_CATEGORIES:-""}
        
        echo
        read -p "Введите ID категорий для отслеживания (Enter для всех): " MONITORED_CATEGORIES
        MONITORED_CATEGORIES=${MONITORED_CATEGORIES:-""}
        
        read -p "Введите ID категорий для игнорирования (Enter для пропуска): " IGNORED_CATEGORIES
        IGNORED_CATEGORIES=${IGNORED_CATEGORIES:-""}
        
        echo
        read -p "Введите домен для webhook (Enter для использования IP): " WEBHOOK_DOMAIN
        WEBHOOK_DOMAIN=${WEBHOOK_DOMAIN:-""}
        
        # Создаем .env файл
        cat > .env << EOF
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN
TELEGRAM_CHAT_ID=$TELEGRAM_CHAT_ID
TELEGRAM_THREAD_ID=$TELEGRAM_THREAD_ID

# Webhook Configuration
WEBHOOK_SECRET=$WEBHOOK_SECRET
WEBHOOK_PORT=$WEBHOOK_PORT
WEBHOOK_PATH=$WEBHOOK_PATH

# AI Configuration
OPENAI_API_KEY=$OPENAI_API_KEY
OPENAI_MODEL=$OPENAI_MODEL

# Base URL for your forum
BASE_URL=$BASE_URL

# Webhook domain (optional, if empty will use server IP)
WEBHOOK_DOMAIN=$WEBHOOK_DOMAIN

# Premium categories (comma-separated category IDs)
PREMIUM_CATEGORIES=$PREMIUM_CATEGORIES

# Categories to monitor (comma-separated category IDs, empty = monitor all)
MONITORED_CATEGORIES=$MONITORED_CATEGORIES

# Categories to ignore (comma-separated category IDs, priority over monitored)
IGNORED_CATEGORIES=$IGNORED_CATEGORIES
EOF
        print_status "Создан файл конфигурации .env"
    else
        print_warning "Конфигурация уже существует"
    fi

    # Docker Compose файл
    print_info "4. Создание Docker Compose файла..."
    cat > docker-compose.yml << EOF
version: '3.8'

services:
  webhook-bot:
    image: $DOCKER_IMAGE
    container_name: webhook_tg_bot
    restart: unless-stopped
    ports:
      - "\${WEBHOOK_PORT:-8080}:8080"
    env_file:
      - .env
    environment:
      - WEBHOOK_PORT=8080
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
EOF
    print_status "Создан файл docker-compose.yml с продакшен образом"

    # Загрузка образа
    print_info "5. Загрузка Docker образа..."
    if docker pull $DOCKER_IMAGE; then
        print_status "Docker образ загружен успешно"
    else
        print_error "Ошибка при загрузке Docker образа"
        print_info "Проверьте доступность образа: $DOCKER_IMAGE"
        return 1
    fi

    # Запуск контейнера
    print_info "6. Запуск контейнера..."
    if $COMPOSE_CMD up -d; then
        print_status "Контейнер запущен успешно"
    else
        print_error "Ошибка при запуске контейнера"
        return 1
    fi

    # Проверка здоровья контейнера
    print_info "7. Проверка состояния контейнера..."
    sleep 10
    if curl -f http://localhost:${WEBHOOK_PORT:-8080}/health &>/dev/null; then
        print_status "Контейнер работает корректно"
    else
        print_warning "Не удалось проверить состояние контейнера"
        print_info "Проверьте логи: $COMPOSE_CMD logs"
    fi

    # Создание systemd сервиса
    print_info "8. Настройка автозапуска..."
    if command -v docker-compose &> /dev/null; then
        SYSTEMD_COMPOSE_CMD="/usr/local/bin/docker-compose"
    else
        SYSTEMD_COMPOSE_CMD="/usr/bin/docker compose"
    fi

    if [[ $EUID -eq 0 ]]; then
        tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << EOF
[Unit]
Description=Webhook Telegram Bot
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=true
WorkingDirectory=$BOT_DIR
ExecStart=$SYSTEMD_COMPOSE_CMD up -d
ExecStop=$SYSTEMD_COMPOSE_CMD down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
        systemctl enable $SERVICE_NAME.service
    else
        sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << EOF
[Unit]
Description=Webhook Telegram Bot
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=true
WorkingDirectory=$BOT_DIR
ExecStart=$SYSTEMD_COMPOSE_CMD up -d
ExecStop=$SYSTEMD_COMPOSE_CMD down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
        sudo systemctl enable $SERVICE_NAME.service
    fi
    print_status "Автозапуск настроен"

    echo
    print_success "Установка завершена успешно!"
    
    # Показать информацию для настройки
    show_webhook_info
}

# Показать информацию о webhook
show_webhook_info() {
    if [ -f "$BOT_DIR/.env" ]; then
        WEBHOOK_PORT=$(grep "^WEBHOOK_PORT=" "$BOT_DIR/.env" | cut -d'=' -f2)
        WEBHOOK_SECRET=$(grep "^WEBHOOK_SECRET=" "$BOT_DIR/.env" | cut -d'=' -f2)
        WEBHOOK_PATH=$(grep "^WEBHOOK_PATH=" "$BOT_DIR/.env" | cut -d'=' -f2)
        WEBHOOK_DOMAIN=$(grep "^WEBHOOK_DOMAIN=" "$BOT_DIR/.env" | cut -d'=' -f2)
        
        if [ -n "$WEBHOOK_DOMAIN" ]; then
            WEBHOOK_URL="http://$WEBHOOK_DOMAIN:$WEBHOOK_PORT$WEBHOOK_PATH"
            CADDY_URL="https://$WEBHOOK_DOMAIN$WEBHOOK_PATH"
        else
            PUBLIC_IP=$(curl -s -4 ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")
            WEBHOOK_URL="http://$PUBLIC_IP:$WEBHOOK_PORT$WEBHOOK_PATH"
            CADDY_URL="https://yourdomain.com$WEBHOOK_PATH"
        fi
        
        echo
        echo -e "${BOLD}📝 Настройка Discourse:${NC}"
        echo -e "   URL: ${CYAN}$WEBHOOK_URL${NC}"
        echo -e "   Secret: ${CYAN}$WEBHOOK_SECRET${NC}"
        echo
        echo -e "${BOLD}🌐 Caddy конфигурация:${NC}"
        if [ -n "$WEBHOOK_DOMAIN" ]; then
            echo -e "   ${CYAN}$WEBHOOK_DOMAIN {${NC}"
            echo -e "   ${CYAN}    handle $WEBHOOK_PATH {${NC}"
            echo -e "   ${CYAN}        reverse_proxy localhost:$WEBHOOK_PORT${NC}"
            echo -e "   ${CYAN}    }${NC}"
            echo -e "   ${CYAN}}${NC}"
            echo -e "   URL для Discourse: ${CYAN}$CADDY_URL${NC}"
        else
            echo -e "   ${CYAN}yourdomain.com {${NC}"
            echo -e "   ${CYAN}    handle $WEBHOOK_PATH {${NC}"
            echo -e "   ${CYAN}        reverse_proxy localhost:$WEBHOOK_PORT${NC}"
            echo -e "   ${CYAN}    }${NC}"
            echo -e "   ${CYAN}}${NC}"
            echo -e "   URL для Discourse: ${CYAN}$CADDY_URL${NC}"
        fi
    fi
}

# Запуск бота
start_bot() {
    print_info "Запуск бота..."
    if [ -d "$BOT_DIR" ]; then
        cd "$BOT_DIR"
        COMPOSE_CMD=$(check_compose_cmd)
        if [ -n "$COMPOSE_CMD" ]; then
            $COMPOSE_CMD up -d
            systemctl start $SERVICE_NAME 2>/dev/null || true
            print_success "Бот запущен"
        else
            print_error "Docker Compose недоступен"
        fi
    else
        print_error "Бот не установлен. Выберите пункт 1 для установки."
    fi
}

# Остановка бота
stop_bot() {
    print_info "Остановка бота..."
    if [ -d "$BOT_DIR" ]; then
        cd "$BOT_DIR"
        COMPOSE_CMD=$(check_compose_cmd)
        if [ -n "$COMPOSE_CMD" ]; then
            $COMPOSE_CMD down
            systemctl stop $SERVICE_NAME 2>/dev/null || true
            print_success "Бот остановлен"
        else
            print_error "Docker Compose недоступен"
        fi
    else
        print_error "Бот не установлен"
    fi
}

# Перезапуск бота
restart_bot() {
    print_info "Перезапуск бота..."
    stop_bot
    sleep 2
    start_bot
}

# Показать логи
show_logs() {
    if [ -d "$BOT_DIR" ]; then
        cd "$BOT_DIR"
        COMPOSE_CMD=$(check_compose_cmd)
        if [ -n "$COMPOSE_CMD" ]; then
            echo -e "${BOLD}📄 Логи бота (Ctrl+C для выхода):${NC}"
            echo
            $COMPOSE_CMD logs -f webhook-bot
        else
            print_error "Docker Compose недоступен"
        fi
    else
        print_error "Бот не установлен"
    fi
}

# Редактирование конфигурации
edit_config() {
    if [ -f "$BOT_DIR/.env" ]; then
        echo -e "${BOLD}📝 Редактирование конфигурации:${NC}"
        echo
        if command -v nano &> /dev/null; then
            nano "$BOT_DIR/.env"
        elif command -v vim &> /dev/null; then
            vim "$BOT_DIR/.env"
        else
            echo "Файл конфигурации: $BOT_DIR/.env"
            echo "Откройте его в любом текстовом редакторе"
        fi
        
        echo
        read -p "Перезапустить бота для применения изменений? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            restart_bot
        fi
    else
        print_error "Конфигурация не найдена. Установите бота сначала."
    fi
}

# Обновление бота
update_bot() {
    print_header
    echo
    print_info "Обновление Webhook Telegram Bot..."
    echo

    if [ ! -d "$BOT_DIR" ]; then
        print_error "Бот не установлен. Установите его сначала."
        return 1
    fi

    cd "$BOT_DIR"
    COMPOSE_CMD=$(check_compose_cmd)

    print_info "1. Остановка текущего контейнера..."
    $COMPOSE_CMD down
    print_status "Контейнер остановлен"

    print_info "2. Загрузка нового образа..."
    if docker pull $DOCKER_IMAGE; then
        print_status "Новый образ загружен"
    else
        print_error "Ошибка при загрузке нового образа"
        print_info "Запускаем старую версию..."
        $COMPOSE_CMD up -d
        return 1
    fi

    print_info "3. Запуск обновленного контейнера..."
    if $COMPOSE_CMD up -d; then
        print_status "Контейнер запущен с новой версией"
    else
        print_error "Ошибка при запуске обновленного контейнера"
        return 1
    fi

    print_info "4. Проверка состояния..."
    sleep 10
    if curl -f http://localhost:${WEBHOOK_PORT:-8080}/health &>/dev/null; then
        print_success "Обновление завершено успешно!"
    else
        print_warning "Контейнер запущен, но состояние неясно"
        print_info "Проверьте логи: $COMPOSE_CMD logs"
    fi
}

# Удаление бота
uninstall_bot() {
    echo -e "${BOLD}${RED}⚠️  ВНИМАНИЕ: Удаление бота${NC}"
    echo
    echo "Это действие удалит:"
    echo "• Все файлы бота в $BOT_DIR"
    echo "• Systemd сервис"
    echo "• Docker контейнеры и образы"
    echo
    read -p "Вы уверены? Введите 'yes' для подтверждения: " -r
    echo
    if [ "$REPLY" = "yes" ]; then
        print_info "Удаление бота..."
        
        # Остановка сервисов
        stop_bot
        
        # Удаление systemd сервиса
        if [[ $EUID -eq 0 ]]; then
            systemctl disable $SERVICE_NAME 2>/dev/null || true
            rm -f /etc/systemd/system/$SERVICE_NAME.service
            systemctl daemon-reload
        else
            sudo systemctl disable $SERVICE_NAME 2>/dev/null || true
            sudo rm -f /etc/systemd/system/$SERVICE_NAME.service
            sudo systemctl daemon-reload
        fi
        
        # Удаление Docker образов
        docker rmi webhook_tg_bot:latest 2>/dev/null || true
        
        # Удаление директории
        if [[ $EUID -eq 0 ]]; then
            rm -rf "$BOT_DIR"
        else
            sudo rm -rf "$BOT_DIR"
        fi
        
        print_success "Бот полностью удален"
    else
        print_info "Удаление отменено"
    fi
}

# Показать помощь
show_help() {
    print_header
    echo
    echo -e "${BOLD}📖 Справка по командам:${NC}"
    echo
    echo -e "   ${CYAN}$SCRIPT_NAME${NC}                  - Интерактивное меню"
    echo -e "   ${CYAN}$SCRIPT_NAME install${NC}          - Установить бота"
    echo -e "   ${CYAN}$SCRIPT_NAME update${NC}           - Обновить бота до последней версии"
    echo -e "   ${CYAN}$SCRIPT_NAME start${NC}            - Запустить бота"
    echo -e "   ${CYAN}$SCRIPT_NAME stop${NC}             - Остановить бота"
    echo -e "   ${CYAN}$SCRIPT_NAME restart${NC}          - Перезапустить бота"
    echo -e "   ${CYAN}$SCRIPT_NAME status${NC}           - Показать статус"
    echo -e "   ${CYAN}$SCRIPT_NAME logs${NC}             - Показать логи"
    echo -e "   ${CYAN}$SCRIPT_NAME uninstall${NC}        - Удалить бота"
    echo -e "   ${CYAN}$SCRIPT_NAME help${NC}             - Эта справка"
    echo
    echo -e "${BOLD}📁 Файлы и директории:${NC}"
    echo -e "   Установка:         ${CYAN}$BOT_DIR${NC}"
    echo -e "   Конфигурация:      ${CYAN}$BOT_DIR/.env${NC}"
    echo -e "   Docker Compose:    ${CYAN}$BOT_DIR/docker-compose.yml${NC}"
    echo -e "   Systemd сервис:    ${CYAN}/etc/systemd/system/$SERVICE_NAME.service${NC}"
    echo -e "   Docker образ:      ${CYAN}$DOCKER_IMAGE${NC}"
    echo
}

# Установка скрипта в систему
install_script() {
    if [ "$0" != "$INSTALL_PATH" ]; then
        print_info "Копирование скрипта в $INSTALL_PATH..."
        if [[ $EUID -eq 0 ]]; then
            cp "$0" "$INSTALL_PATH"
            chmod +x "$INSTALL_PATH"
        else
            sudo cp "$0" "$INSTALL_PATH"
            sudo chmod +x "$INSTALL_PATH"
        fi
        print_success "Скрипт установлен. Используйте команду: $SCRIPT_NAME"
    fi
}

# Основная функция
main() {
    # Устанавливаем скрипт в систему при первом запуске
    install_script
    
    case "${1:-}" in
        "install")
            install_bot
            ;;
        "update")
            update_bot
            ;;
        "start")
            start_bot
            ;;
        "stop")
            stop_bot
            ;;
        "restart")
            restart_bot
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "uninstall")
            uninstall_bot
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        "")
            # Интерактивное меню
            while true; do
                show_menu
                read -n 1 -r choice
                echo
                echo
                
                case $choice in
                    1) install_bot ;;
                    2) update_bot ;;
                    3) start_bot ;;
                    4) stop_bot ;;
                    5) restart_bot ;;
                    6) show_logs ;;
                    7) continue ;;
                    8) edit_config ;;
                    9) uninstall_bot ;;
                    0) 
                        echo -e "${BOLD}До свидания! 👋${NC}"
                        exit 0 
                        ;;
                    *)
                        print_error "Неверный выбор. Попробуйте снова."
                        ;;
                esac
                
                if [ "$choice" != "7" ] && [ "$choice" != "6" ]; then
                    echo
                    read -p "Нажмите Enter для продолжения..." -r
                fi
            done
            ;;
        *)
            echo -e "${RED}Неизвестная команда: $1${NC}"
            echo "Используйте '$SCRIPT_NAME help' для справки"
            exit 1
            ;;
    esac
}

# Запуск
main "$@"
