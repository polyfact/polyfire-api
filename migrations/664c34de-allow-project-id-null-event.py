def migrate(cur, rls=False):
    cur.execute("""
        ALTER TABLE events ALTER COLUMN project_id DROP NOT NULL;
        ALTER TABLE models ALTER COLUMN tags SET DEFAULT array[]::text[];
    """)

def rollback(cur, rls=False):
    cur.execute("""
        ALTER TABLE events ALTER COLUMN project_id SET NOT NULL;
        ALTER TABLE models ALTER COLUMN tags DROP DEFAULT;
    """)
