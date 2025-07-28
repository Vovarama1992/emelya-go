-- Добавляем обратно block_until
ALTER TABLE deposits ADD COLUMN block_until TIMESTAMPTZ;

-- Восстанавливаем по block_days
UPDATE deposits
SET block_until = now() + (block_days || ' days')::interval
WHERE block_days IS NOT NULL;

-- Удаляем block_days
ALTER TABLE deposits DROP COLUMN block_days;