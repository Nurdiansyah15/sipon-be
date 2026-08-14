ALTER TABLE user_identities DROP CONSTRAINT IF EXISTS user_identities_kind_check;
ALTER TABLE user_identities ADD CONSTRAINT user_identities_kind_check
    CHECK (kind IN ('EMAIL', 'PHONE', 'USERNAME', 'NIS'));
