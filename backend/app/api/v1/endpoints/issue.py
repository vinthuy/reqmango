"""
Issue API Endpoints - 工作项管理接口
"""
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.models.project import Project
from app.schemas.issue import (
    IssueCreate,
    IssueUpdate,
    IssueResponse,
    IssueLite,
    IssueSearchResult,
    IssuePriority
)
from app.services import issue as issue_service
from app.core.exceptions import NotFoundException, ValidationException

router = APIRouter()


# ==================== Issue CRUD ====================

@router.post("/", response_model=IssueResponse, status_code=201)
async def create_issue(
    project_id: int,
    workspace_id: int,
    issue_data: IssueCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建工作项
    
    在指定项目中创建一个新的工作项（Issue、Task、Bug等）。
    可以设置名称、描述、优先级、状态、负责人、标签等属性。
    """
    # 验证项目访问权限
    project = await db.get(Project, project_id)
    if not project or project.is_deleted:
        raise HTTPException(status_code=404, detail="Project not found")
    
    issue = await issue_service.create_issue(
        db=db,
        issue_data=issue_data,
        project_id=project_id,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    
    return issue_service.build_issue_response(issue)


@router.get("/", response_model=List[IssueResponse])
async def list_issues(
    project_id: int,
    workspace_id: int,
    state_id: Optional[int] = None,
    priority: Optional[IssuePriority] = None,
    assignee_id: Optional[int] = None,
    parent_id: Optional[int] = None,
    cycle_id: Optional[int] = None,
    module_id: Optional[int] = None,
    search: Optional[str] = None,
    is_draft: Optional[bool] = None,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出项目的工作项
    
    支持多种筛选条件：
    - state_id: 按状态筛选
    - priority: 按优先级筛选
    - assignee_id: 按负责人筛选
    - parent_id: 按父工作项筛选
    - cycle_id: 按周期筛选
    - module_id: 按模块筛选
    - search: 搜索关键词
    - is_draft: 是否为草稿
    """
    issues = await issue_service.list_project_issues(
        db=db,
        project_id=project_id,
        state_id=state_id,
        priority=priority,
        assignee_id=assignee_id,
        parent_id=parent_id,
        cycle_id=cycle_id,
        module_id=module_id,
        search=search,
        is_draft=is_draft,
        limit=limit,
        offset=offset
    )
    
    return [issue_service.build_issue_response(issue) for issue in issues]


@router.get("/{issue_id}", response_model=IssueResponse)
async def get_issue(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作项详情
    
    返回工作项的完整信息，包括关联的项目、状态、自定义字段值等。
    """
    issue = await issue_service.get_issue_by_id(db, issue_id)
    return issue_service.build_issue_response(issue)


@router.put("/{issue_id}", response_model=IssueResponse)
async def update_issue(
    issue_id: int,
    update_data: IssueUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新工作项
    
    可以更新工作项的名称、描述、优先级、状态、负责人、标签等属性。
    每次更新都会记录活动历史。
    """
    issue = await issue_service.update_issue(
        db=db,
        issue_id=issue_id,
        update_data=update_data,
        user_id=current_user.id
    )
    return issue_service.build_issue_response(issue)


@router.delete("/{issue_id}", status_code=204)
async def delete_issue(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除工作项（软删除）
    
    工作项不会真正删除，而是标记为已删除状态。
    可以通过恢复接口恢复。
    """
    await issue_service.delete_issue(db, issue_id)
    return None


@router.post("/{issue_id}/archive", status_code=204)
async def archive_issue(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    归档工作项
    
    归档的工作项会从常规列表中隐藏，但仍可通过归档列表查看。
    """
    await issue_service.archive_issue(db, issue_id)
    return None


@router.post("/{issue_id}/restore", response_model=IssueResponse)
async def restore_issue(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    恢复工作项
    
    恢复已删除或归档的工作项。
    """
    issue = await issue_service.restore_issue(db, issue_id)
    return issue_service.build_issue_response(issue)


# ==================== Issue Activities ====================

@router.get("/{issue_id}/activities", response_model=List[dict])
async def get_issue_activities(
    issue_id: int,
    limit: int = Query(50, ge=1, le=100),
    offset: int = Query(0, ge=0),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作项活动历史
    
    返回工作项的所有变更记录，包括创建、更新、状态变更等。
    """
    activities = await issue_service.get_issue_activities(
        db=db,
        issue_id=issue_id,
        limit=limit,
        offset=offset
    )
    
    return [
        {
            "id": activity.id,
            "issue_id": activity.issue_id,
            "verb": activity.verb,
            "field": activity.field,
            "old_value": activity.old_value,
            "new_value": activity.new_value,
            "comment": activity.comment,
            "actor_id": activity.actor_id,
            "created_at": activity.created_at
        }
        for activity in activities
    ]


# ==================== Issue Statistics ====================

@router.get("/statistics", response_model=dict)
async def get_issue_statistics(
    project_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取项目工作项统计
    
    返回工作项的总数、按状态分组统计、按优先级分组统计等。
    """
    stats = await issue_service.get_issue_statistics(db, project_id)
    return stats


# ==================== Issue Search ====================

@router.get("/search", response_model=List[IssueSearchResult])
async def search_issues(
    workspace_id: int,
    query: str = Query(..., min_length=1),
    project_id: Optional[int] = None,
    limit: int = Query(10, ge=1, le=50),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    搜索工作项
    
    在工作空间或项目范围内搜索工作项，支持名称和描述搜索。
    """
    # TODO: 实现跨项目搜索
    if project_id:
        issues = await issue_service.list_project_issues(
            db=db,
            project_id=project_id,
            search=query,
            limit=limit
        )
        
        return [
            IssueSearchResult(
                id=issue.id,
                name=issue.name,
                sequence_id=issue.sequence_id,
                project_identifier=issue.project.identifier if issue.project else "",
                project_id=issue.project_id,
                workspace_slug=workspace_id  # TODO: 获取 workspace slug
            )
            for issue in issues
        ]
    
    return []


# ==================== Bulk Operations ====================

@router.post("/bulk/update", response_model=List[IssueResponse])
async def bulk_update_issues(
    project_id: int,
    issue_ids: List[int],
    update_data: IssueUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    批量更新工作项
    
    同时更新多个工作项的属性（如状态、优先级、负责人等）。
    """
    results = []
    for issue_id in issue_ids:
        issue = await issue_service.update_issue(
            db=db,
            issue_id=issue_id,
            update_data=update_data,
            user_id=current_user.id
        )
        results.append(issue_service.build_issue_response(issue))
    
    return results


@router.post("/bulk/delete", status_code=204)
async def bulk_delete_issues(
    issue_ids: List[int],
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    批量删除工作项
    """
    for issue_id in issue_ids:
        await issue_service.delete_issue(db, issue_id)
    
    return None


# ==================== Assignee Management ====================

@router.post("/{issue_id}/assignees", response_model=dict)
async def add_issue_assignee(
    issue_id: int,
    user_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    添加工作项负责人
    
    将指定用户添加为工作项的负责人。
    """
    result = await issue_service.add_assignee(
        db=db,
        issue_id=issue_id,
        user_id=user_id,
        actor_id=current_user.id
    )
    return result


@router.delete("/{issue_id}/assignees/{user_id}", response_model=dict)
async def remove_issue_assignee(
    issue_id: int,
    user_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    移除工作项负责人
    
    从工作项的负责人列表中移除指定用户。
    """
    result = await issue_service.remove_assignee(
        db=db,
        issue_id=issue_id,
        user_id=user_id,
        actor_id=current_user.id
    )
    return result


# ==================== Label Management ====================

@router.post("/{issue_id}/labels", response_model=dict)
async def add_issue_label(
    issue_id: int,
    label_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    添加工作项标签
    
    将指定标签添加到工作项。
    """
    result = await issue_service.add_label(
        db=db,
        issue_id=issue_id,
        label_id=label_id,
        actor_id=current_user.id
    )
    return result


@router.delete("/{issue_id}/labels/{label_id}", response_model=dict)
async def remove_issue_label(
    issue_id: int,
    label_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    移除工作项标签
    
    从工作项的标签列表中移除指定标签。
    """
    result = await issue_service.remove_label(
        db=db,
        issue_id=issue_id,
        label_id=label_id,
        actor_id=current_user.id
    )
    return result


# ==================== Cycle Management ====================

@router.post("/{issue_id}/cycle", response_model=dict)
async def set_issue_cycle(
    issue_id: int,
    cycle_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    设置工作项周期
    
    将工作项添加到指定的迭代周期。
    """
    result = await issue_service.set_issue_cycle(
        db=db,
        issue_id=issue_id,
        cycle_id=cycle_id,
        actor_id=current_user.id
    )
    return result


@router.delete("/{issue_id}/cycle", response_model=dict)
async def remove_issue_cycle(
    issue_id: int,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    移除工作项周期
    
    将工作项从当前周期中移除。
    """
    result = await issue_service.remove_issue_cycle(
        db=db,
        issue_id=issue_id,
        actor_id=current_user.id
    )
    return result