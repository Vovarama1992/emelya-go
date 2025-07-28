-- Возвращаем block_until
ALTER TABLE tariffs ADD COLUMN block_until TIMESTAMPTZ;

-- Пересчитываем обратно: now() + block_days
UPDATE tariffs SET block_until = now() + (block_days || ' days')::interval;

-- Удаляем block_days
ALTER TABLE tariffs DROP COLUMN block_days;