-- Option 1: Add a column to store the constraint name
DO $$ 
BEGIN
  ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_owner_fkey;
END $$; 