-- Добавляем поле block_days
ALTER TABLE deposits ADD COLUMN block_days INT;

-- Переносим значение из block_until
UPDATE deposits
SET block_days = EXTRACT(DAY FROM block_until - now())
WHERE block_until IS NOT NULL;

-- Удаляем block_until
ALTER TABLE deposits DROP COLUMN block_until;