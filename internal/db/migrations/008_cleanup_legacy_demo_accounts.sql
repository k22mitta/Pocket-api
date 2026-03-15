-- Before the CSV-import fix, real (non-demo) users who uploaded a CSV or hit
-- an empty-account path got a hardcoded "Demo Checking" / "Demo Bank" account
-- (access_token='demo', a legacy sentinel only ever written by the old
-- CreateDemoAccount helper). Real users must never see demo artifacts.
-- Deleting the plaid_items row cascades to its accounts, which cascades to
-- their transactions (see migrations 003/004).
DELETE FROM plaid_items
WHERE access_token = 'demo'
  AND user_id NOT IN (SELECT id FROM users WHERE email = 'demo@example.com');
