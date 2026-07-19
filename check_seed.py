import psycopg2

conn = psycopg2.connect(host='localhost', port=5432, dbname='reqmango', user='postgres', password='postgres')
cur = conn.cursor()

cur.execute("SELECT count(*) FROM issues")
print("Issues:", cur.fetchone()[0])

cur.execute("SELECT count(*) FROM automation_rules")
print("Rules:", cur.fetchone()[0])

cur.execute("SELECT count(*) FROM automation_executions")
print("Executions:", cur.fetchone()[0])

cur.execute("SELECT count(*) FROM issues WHERE parent_id IS NOT NULL")
print("ChildIssues:", cur.fetchone()[0])

cur.execute("SELECT id, name, trigger_type, is_enabled FROM automation_rules ORDER BY id LIMIT 20")
for r in cur.fetchall():
    print("  #{0} {1} trigger={2} enabled={3}".format(r[0], r[1], r[2], r[3]))

# Find some parent-child pairs to test
cur.execute("""
    SELECT p.id as pid, p.name as pname, p.state_id as pstate,
           c.id as cid, c.name as cname, c.state_id as cstate
    FROM issues c
    JOIN issues p ON p.id = c.parent_id
    LIMIT 10
""")
print("\nParent-Child pairs:")
for r in cur.fetchall():
    print("  P#{0} [{1}] state={2} <- C#{3} [{4}] state={5}".format(r[0], r[1], r[2], r[3], r[4], r[5]))

# States
cur.execute("SELECT id, name, \"group\" FROM states WHERE project_id = 1 AND is_active = true ORDER BY id")
print("\nStates (project 1):")
for r in cur.fetchall():
    print("  id={0} name='{1}' group={2}".format(r[0], r[1], r[2]))

conn.close()
