"""
Module Services - 模块管理业务逻辑层
"""
from typing import Optional, List, Dict, Any
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload
from sqlalchemy import and_, func

from app.models.module import Module
from app.models.issue import Issue, IssueModule
from app.models.project import Project
from app.models.state import State
from app.schemas.module import (
    ModuleCreate,
    ModuleUpdate,
    ModuleResponse
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== Module CRUD ====================

async def create_module(
    db: AsyncSession,
    module_data: ModuleCreate,
    workspace_id: int,
    user_id: int
) -> Module:
    """创建模块"""
    # 验证项目存在
    project = await db.get(Project, module_data.project_id)
    if not project or project.is_deleted:
        raise NotFoundException("Project not found")
    
    # 创建模块
    module = Module(
        name=module_data.name,
        description=module_data.description,
        target_date=module_data.target_date,
        project_id=module_data.project_id,
        workspace_id=workspace_id,
        created_by_id=user_id
    )
    
    db.add(module)
    await db.commit()
    await db.refresh(module)
    
    return module


async def get_module_by_id(db: AsyncSession, module_id: int) -> Module:
    """获取模块详情"""
    result = await db.execute(
        select(Module)
        .where(Module.id == module_id, Module.is_deleted == False)
        .options(
            selectinload(Module.project),
            selectinload(Module.parent),
            selectinload(Module.sub_modules)
        )
    )
    module = result.scalar_one_or_none()
    if not module:
        raise NotFoundException("Module not found")
    return module


async def list_project_modules(
    db: AsyncSession,
    project_id: int,
    parent_id: Optional[int] = None,
    include_archived: bool = False,
    limit: int = 50,
    offset: int = 0
) -> List[Module]:
    """列出项目的模块"""
    query = select(Module).where(
        Module.project_id == project_id,
        Module.is_deleted == False
    )
    
    # 父模块筛选
    if parent_id:
        query = query.where(Module.parent_id == parent_id)
    
    # 排序
    query = query.order_by(Module.sequence, Module.created_at.desc())
    
    # 分页
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query.options(
        selectinload(Module.project),
        selectinload(Module.sub_modules)
    ))
    
    return list(result.scalars().all())


async def update_module(
    db: AsyncSession,
    module_id: int,
    update_data: ModuleUpdate,
    user_id: int
) -> Module:
    """更新模块"""
    module = await get_module_by_id(db, module_id)
    
    # 更新字段
    if update_data.name is not None:
        module.name = update_data.name
    if update_data.description is not None:
        module.description = update_data.description
    if update_data.target_date is not None:
        module.target_date = update_data.target_date
    
    await db.commit()
    await db.refresh(module)
    
    return module


async def delete_module(db: AsyncSession, module_id: int):
    """删除模块（软删除）"""
    module = await get_module_by_id(db, module_id)
    module.is_deleted = True
    await db.commit()


# ==================== Module Issue Management ====================

async def add_issue_to_module(
    db: AsyncSession,
    module_id: int,
    issue_id: int,
    user_id: int
) -> Dict[str, Any]:
    """将工作项添加到模块"""
    # 验证模块存在
    module = await get_module_by_id(db, module_id)
    
    # 验证工作项存在
    result = await db.execute(
        select(Issue).where(Issue.id == issue_id, Issue.is_deleted == False)
    )
    issue = result.scalar_one_or_none()
    if not issue:
        raise NotFoundException("Issue not found")
    
    # 验证工作项属于同一项目
    if issue.project_id != module.project_id:
        raise ValidationException("Issue does not belong to this module's project")
    
    # 检查是否已在模块中
    existing_result = await db.execute(
        select(IssueModule).where(
            IssueModule.module_id == module_id,
            IssueModule.issue_id == issue_id
        )
    )
    existing = existing_result.scalar_one_or_none()
    if existing:
        raise ValidationException("Issue is already in this module")
    
    # 创建关联
    association = IssueModule(
        module_id=module_id,
        issue_id=issue_id
    )
    
    db.add(association)
    await db.commit()
    
    return {
        "module_id": module_id,
        "issue_id": issue_id,
        "action": "added"
    }


async def remove_issue_from_module(
    db: AsyncSession,
    module_id: int,
    issue_id: int,
    user_id: int
) -> Dict[str, Any]:
    """从模块移除工作项"""
    # 验证模块存在
    module = await get_module_by_id(db, module_id)
    
    # 查找关联
    result = await db.execute(
        select(IssueModule).where(
            IssueModule.module_id == module_id,
            IssueModule.issue_id == issue_id
        )
    )
    association = result.scalar_one_or_none()
    
    if not association:
        raise ValidationException("Issue is not in this module")
    
    # 删除关联
    await db.delete(association)
    await db.commit()
    
    return {
        "module_id": module_id,
        "issue_id": issue_id,
        "action": "removed"
    }


async def get_module_issues(
    db: AsyncSession,
    module_id: int,
    state_id: Optional[int] = None,
    priority: Optional[str] = None,
    limit: int = 50,
    offset: int = 0
) -> List[Issue]:
    """获取模块内的工作项"""
    # 验证模块存在
    module = await get_module_by_id(db, module_id)
    
    # 查询工作项（通过IssueModule关联）
    query = select(Issue).join(IssueModule).where(
        IssueModule.module_id == module_id,
        Issue.is_deleted == False
    )
    
    # 状态筛选
    if state_id:
        query = query.where(Issue.state_id == state_id)
    
    # 优先级筛选
    if priority:
        query = query.where(Issue.priority == priority)
    
    # 排序
    query = query.order_by(Issue.sort_order, Issue.sequence_id.desc())
    
    # 分页
    query = query.limit(limit).offset(offset)
    
    result = await db.execute(query.options(
        selectinload(Issue.project),
        selectinload(Issue.state)
    ))
    
    return list(result.scalars().all())


# ==================== Module Progress & Statistics ====================

async def get_module_progress(
    db: AsyncSession,
    module_id: int
) -> Dict[str, Any]:
    """获取模块进度"""
    module = await get_module_by_id(db, module_id)
    
    # 统计总数
    total_result = await db.execute(
        select(func.count(Issue.id))
        .join(IssueModule)
        .where(
            IssueModule.module_id == module_id,
            Issue.is_deleted == False
        )
    )
    total_issues = total_result.scalar_one_or_none() or 0
    
    # 统计已完成数（通过state group判断）
    completed_result = await db.execute(
        select(func.count(Issue.id))
        .join(IssueModule)
        .join(State, Issue.state_id == State.id)
        .where(
            IssueModule.module_id == module_id,
            Issue.is_deleted == False,
            State.group == "done"
        )
    )
    completed_issues = completed_result.scalar_one_or_none() or 0
    
    # 计算进度
    progress = (completed_issues / total_issues * 100) if total_issues > 0 else 0
    
    # 按状态分组统计
    state_stats_result = await db.execute(
        select(State.name, State.group, func.count(Issue.id))
        .join(IssueModule)
        .join(State, Issue.state_id == State.id)
        .where(
            IssueModule.module_id == module_id,
            Issue.is_deleted == False
        )
        .group_by(State.name, State.group)
    )
    state_stats = [
        {"state": row[0], "group": row[1], "count": row[2]}
        for row in state_stats_result.all()
    ]
    
    return {
        "module_id": module_id,
        "module_name": module.name,
        "total_issues": total_issues,
        "completed_issues": completed_issues,
        "progress": round(progress, 2),
        "state_breakdown": state_stats
    }


async def get_module_statistics(
    db: AsyncSession,
    module_id: int
) -> Dict[str, Any]:
    """获取模块详细统计"""
    progress = await get_module_progress(db, module_id)
    module = await get_module_by_id(db, module_id)
    
    # 优先级统计
    priority_result = await db.execute(
        select(Issue.priority, func.count(Issue.id))
        .join(IssueModule)
        .where(
            IssueModule.module_id == module_id,
            Issue.is_deleted == False
        )
        .group_by(Issue.priority)
    )
    priority_stats = {row[0]: row[1] for row in priority_result.all()}
    
    # 工作项统计
    issue_stats_result = await db.execute(
        select(
            func.count(Issue.id).label('total'),
            func.count(Issue.start_date).label('with_start'),
            func.count(Issue.target_date).label('with_target'),
        )
        .join(IssueModule)
        .where(
            IssueModule.module_id == module_id,
            Issue.is_deleted == False
        )
    )
    issue_stats = issue_stats_result.first()
    
    return {
        **progress,
        "priority_breakdown": priority_stats,
        "issue_stats": {
            "total": issue_stats.total if issue_stats else 0,
            "with_start_date": issue_stats.with_start if issue_stats else 0,
            "with_target_date": issue_stats.with_target if issue_stats else 0
        },
        "target_date": str(module.target_date) if module.target_date else None,
        "has_sub_modules": len(module.sub_modules) > 0
    }


# ==================== Module Hierarchy ====================

async def get_module_tree(
    db: AsyncSession,
    project_id: int
) -> List[Dict[str, Any]]:
    """获取模块树形结构"""
    # 获取所有顶级模块（无父模块）
    result = await db.execute(
        select(Module).where(
            Module.project_id == project_id,
            Module.is_deleted == False,
            Module.parent_id == None
        ).options(selectinload(Module.sub_modules))
    )
    
    top_modules = list(result.scalars().all())
    
    # 构建树形结构
    tree = []
    for module in top_modules:
        node = await _build_module_tree_node(db, module)
        tree.append(node)
    
    return tree


async def _build_module_tree_node(
    db: AsyncSession,
    module: Module
) -> Dict[str, Any]:
    """构建模块树节点"""
    progress = await get_module_progress(db, module.id)
    
    node = {
        "id": module.id,
        "name": module.name,
        "description": module.description,
        "sequence": module.sequence,
        "progress": progress.get('progress', 0),
        "total_issues": progress.get('total_issues', 0),
        "completed_issues": progress.get('completed_issues', 0),
        "children": []
    }
    
    # 递归构建子模块
    for sub_module in module.sub_modules:
        if not sub_module.is_deleted:
            child_node = await _build_module_tree_node(db, sub_module)
            node["children"].append(child_node)
    
    return node


# ==================== Helper Functions ====================

def build_module_response(module: Module, total_issues: int = 0, completed_issues: int = 0) -> Dict[str, Any]:
    """构建Module Response"""
    progress = (completed_issues / total_issues * 100) if total_issues > 0 else 0
    
    return {
        "id": module.id,
        "name": module.name,
        "description": module.description,
        "sequence": module.sequence,
        "target_date": module.target_date,
        "progress": round(progress, 2),
        "total_issues": total_issues,
        "completed_issues": completed_issues,
        "project_id": module.project_id,
        "workspace_id": module.workspace_id,
        "parent_id": module.parent_id if module.parent_id else None,
        "created_at": module.created_at,
        "updated_at": module.updated_at,
    }