ALTER TABLE deposits
    ALTER COLUMN daily_reward TYPE NUMERIC(12, 6);

ALTER TABLE tariffs
    ALTER COLUMN daily_reward TYPE NUMERIC(12, 6);