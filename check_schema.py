import psycopg2
c = psycopg2.connect(host='localhost', port=5432, dbname='reqmango', user='postgres', password='postgres')
cur = c.cursor()

# Check if resolved_at column exists
cur.execute("SELECT column_name FROM information_schema.columns WHERE table_name='issues' AND column_name IN ('resolved_at','completed_at','closed_at')")
print('Columns found:', [r[0] for r in cur.fetchall()])

# Check issue schema
cur.execute("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='issues' ORDER BY ordinal_position")
print('\nIssue columns:')
for r in cur.fetchall():
    print(f'  {r[0]:30s} {r[1]}')

c.close()
