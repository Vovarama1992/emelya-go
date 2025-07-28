ALTER TABLE deposits
    ALTER COLUMN daily_reward TYPE NUMERIC(12, 2);

ALTER TABLE tariffs
    ALTER COLUMN daily_reward TYPE NUMERIC(12, 2);