#!/bin/sh

echo "Ожидаем Postgres..."
until pg_isready -h db -p 5432 -U emelya; do
  sleep 1
done

echo "Запускаем приложение..."
exec /app/main "$@"