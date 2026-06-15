import asyncio
import asyncpg

async def test():
    try:
        # 先连接到 postgres 数据库
        conn = await asyncpg.connect(
            user='postgres',
            password='',
            database='postgres',
            host='localhost',
            port=5432
        )
        print('Connected to postgres database!')
        
        # 检查 reqmanpy 数据库是否存在
        result = await conn.fetchval(
            "SELECT 1 FROM pg_database WHERE datname = 'reqmanpy'"
        )
        
        if result:
            print('Database reqmanpy already exists')
        else:
            print('Creating reqmanpy database...')
            await conn.execute('CREATE DATABASE reqmanpy')
            print('Database reqmanpy created successfully!')
        
        await conn.close()
        
        # 连接到 reqmanpy 数据库
        print('\nConnecting to reqmanpy database...')
        reqman_conn = await asyncpg.connect(
            user='postgres',
            password='',
            database='reqmanpy',
            host='localhost',
            port=5432
        )
        print('Connected to reqmanpy database!')
        await reqman_conn.close()
        print('\nAll tests passed!')
        
    except Exception as e:
        print(f'Failed: {e}')

asyncio.run(test())