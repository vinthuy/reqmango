"""
Database Initialization Script
使用SQLAlchemy创建所有数据库表
"""
import asyncio
import sys
from pathlib import Path

# 添加backend到path
sys.path.insert(0, str(Path(__file__).parent.parent))

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

from app.models.base import Base
from app.models import (
    workspace, project, user, state, issue_type,
    issue, label, cycle, module, custom_field,
    workflow, estimate_point
)

# 数据库URL - 根据环境变量或默认值
DATABASE_URL = "sqlite+aiosqlite:///./test.db"  # 使用SQLite作为示例

# PostgreSQL示例:
# DATABASE_URL = "postgresql+asyncpg://user:password@localhost:5432/reqmanpy"


async def create_tables():
    """创建所有表"""
    engine = create_async_engine(DATABASE_URL, echo=True)

    async with engine.begin() as conn:
        # 删除所有表（仅用于开发环境）
        # await conn.run_sync(Base.metadata.drop_all)

        # 创建所有表
        await conn.run_sync(Base.metadata.create_all)

    print("所有表创建成功!")

    await engine.dispose()


async def init_database():
    """初始化数据库"""
    engine = create_async_engine(DATABASE_URL, echo=True)

    async with engine.begin() as conn:
        # 测试连接
        await conn.execute(text("SELECT 1"))
        print("数据库连接成功!")

        # 创建所有表
        await conn.run_sync(Base.metadata.create_all)
        print("所有表创建成功!")

    await engine.dispose()


def main():
    """主函数"""
    print("开始初始化数据库...")
    print(f"数据库URL: {DATABASE_URL}")

    try:
        asyncio.run(init_database())
        print("\n数据库初始化完成!")
    except Exception as e:
        print(f"\n数据库初始化失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
