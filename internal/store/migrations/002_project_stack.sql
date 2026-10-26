-- The kind of app found in a project's repository (F2-03), such as "nextjs" or "go".
ALTER TABLE projects ADD COLUMN stack TEXT NOT NULL DEFAULT '';
