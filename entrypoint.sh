#!/bin/sh

# Устанавливаем права на файлы
chmod +x /app/entrypoint.sh
chmod +x /app/main

# Ждём, пока БД будет готова
echo "Ждём Postgres..."
until pg_isready -h db -p 5432 -U emelya; do
  sleep 1
done

echo "Выполняем миграции..."
migrate -path ./migrations -database "$DATABASE_URL?sslmode=disable" up

echo "Запускаем приложение..."

# Попробуем запустить приложение, если оно есть
if [ -f /app/main ]; then
  exec /app/main "$@"
else
  echo "Бинарник не найден, контейнер будет жить..."
  tail -f /dev/null
fi