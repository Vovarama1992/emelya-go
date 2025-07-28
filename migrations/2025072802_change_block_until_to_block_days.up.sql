-- Добавляем временное поле
ALTER TABLE tariffs ADD COLUMN block_days INT;

-- Переносим данные: считаем разницу между block_until и now()
UPDATE tariffs SET block_days = EXTRACT(DAY FROM block_until - now());

-- Удаляем старое поле
ALTER TABLE tariffs DROP COLUMN block_until;
