"""
Workflow API Endpoints - 工作流和自动化接口
"""
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.schemas.workflow import (
    StateTransitionCreate,
    StateTransitionUpdate,
    StateTransitionResponse,
    AutomationRuleCreate,
    AutomationRuleUpdate,
    AutomationRuleResponse,
    AutomationRuleLite,
    AutomationExecutionLogResponse,
    AutomationTemplateResponse
)
from app.services import workflow as workflow_service
from app.core.exceptions import NotFoundException

router = APIRouter()


# ==================== State Transition Endpoints ====================

@router.post("/transitions", response_model=StateTransitionResponse, status_code=201)
async def create_state_transition(
    project_id: UUID,
    workspace_id: UUID,
    transition_data: StateTransitionCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建状态转换规则
    
    定义哪些状态可以转换到哪些状态。
    """
    transition_data.project_id = project_id
    
    transition = await workflow_service.create_state_transition(
        db=db,
        transition_data=transition_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return transition


@router.get("/transitions", response_model=List[StateTransitionResponse])
async def list_state_transitions(
    project_id: UUID,
    source_state_id: Optional[UUID] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出状态转换规则
    """
    transitions = await workflow_service.get_project_state_transitions(
        db=db,
        project_id=project_id,
        source_state_id=source_state_id
    )
    return transitions


@router.get("/transitions/{transition_id}", response_model=StateTransitionResponse)
async def get_state_transition(
    transition_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取状态转换规则详情
    """
    transition = await workflow_service.get_state_transition_by_id(db, transition_id)
    return transition


@router.put("/transitions/{transition_id}", response_model=StateTransitionResponse)
async def update_state_transition(
    transition_id: UUID,
    update_data: StateTransitionUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新状态转换规则
    """
    transition = await workflow_service.update_state_transition(
        db=db,
        transition_id=transition_id,
        update_data=update_data
    )
    return transition


@router.delete("/transitions/{transition_id}", status_code=204)
async def delete_state_transition(
    transition_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除状态转换规则（软删除）
    """
    await workflow_service.delete_state_transition(db, transition_id)
    return None


@router.get("/transitions/can-transition", response_model=bool)
async def check_can_transition(
    project_id: UUID,
    source_state_id: UUID,
    target_state_id: UUID,
    issue_type_id: Optional[UUID] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    检查是否允许状态转换
    """
    can_transition = await workflow_service.can_transition(
        db=db,
        project_id=project_id,
        source_state_id=source_state_id,
        target_state_id=target_state_id,
        issue_type_id=issue_type_id
    )
    return can_transition


# ==================== Automation Rule Endpoints ====================

@router.post("/automations", response_model=AutomationRuleResponse, status_code=201)
async def create_automation_rule(
    project_id: UUID,
    workspace_id: UUID,
    rule_data: AutomationRuleCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建自动化规则
    
    包含触发器、条件和动作配置。
    """
    rule_data.project_id = project_id
    
    rule = await workflow_service.create_automation_rule(
        db=db,
        rule_data=rule_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return rule


@router.get("/automations", response_model=List[AutomationRuleResponse])
async def list_automation_rules(
    project_id: UUID,
    enabled_only: bool = False,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出自动化规则
    """
    rules = await workflow_service.get_project_automation_rules(
        db=db,
        project_id=project_id,
        enabled_only=enabled_only
    )
    return rules


@router.get("/automations/lite", response_model=List[AutomationRuleLite])
async def list_automation_rules_lite(
    project_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出自动化规则（轻量版）
    """
    rules = await workflow_service.get_project_automation_rules(
        db=db,
        project_id=project_id
    )
    return rules


@router.get("/automations/{rule_id}", response_model=AutomationRuleResponse)
async def get_automation_rule(
    rule_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取自动化规则详情
    """
    rule = await workflow_service.get_automation_rule_by_id(db, rule_id)
    return rule


@router.put("/automations/{rule_id}", response_model=AutomationRuleResponse)
async def update_automation_rule(
    rule_id: UUID,
    update_data: AutomationRuleUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新自动化规则
    """
    rule = await workflow_service.update_automation_rule(
        db=db,
        rule_id=rule_id,
        update_data=update_data
    )
    return rule


@router.delete("/automations/{rule_id}", status_code=204)
async def delete_automation_rule(
    rule_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除自动化规则（软删除）
    """
    await workflow_service.delete_automation_rule(db, rule_id)
    return None


@router.post("/automations/{rule_id}/toggle", response_model=AutomationRuleResponse)
async def toggle_automation_rule(
    rule_id: UUID,
    enabled: bool,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    启用/禁用自动化规则
    """
    rule = await workflow_service.toggle_automation_rule(
        db=db,
        rule_id=rule_id,
        enabled=enabled
    )
    return rule


# ==================== Automation Execution Log Endpoints ====================

@router.get("/automations/{rule_id}/logs", response_model=List[AutomationExecutionLogResponse])
async def list_automation_logs(
    rule_id: UUID,
    limit: int = Query(default=50, le=200),
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取自动化规则执行日志
    """
    logs = await workflow_service.get_rule_execution_logs(
        db=db,
        rule_id=rule_id,
        limit=limit
    )
    return logs


# ==================== Automation Template Endpoints ====================

@router.get("/templates", response_model=List[AutomationTemplateResponse])
async def list_automation_templates(
    category: Optional[str] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出自动化模板
    """
    templates = await workflow_service.get_all_templates(
        db=db,
        category=category
    )
    return templates


@router.post("/templates/{template_id}/apply", response_model=AutomationRuleResponse)
async def apply_automation_template(
    template_id: UUID,
    project_id: UUID,
    workspace_id: UUID,
    name_override: Optional[str] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    从模板创建自动化规则
    """
    rule = await workflow_service.create_rule_from_template(
        db=db,
        template_id=template_id,
        project_id=project_id,
        workspace_id=workspace_id,
        user_id=current_user.id,
        name_override=name_override
    )
    return rule


# ==================== Workflow Validation ====================

@router.post("/validate-transition")
async def validate_transition(
    project_id: UUID,
    source_state_id: UUID,
    target_state_id: UUID,
    issue_type_id: Optional[UUID] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    验证状态转换是否有效
    """
    can_transition = await workflow_service.can_transition(
        db=db,
        project_id=project_id,
        source_state_id=source_state_id,
        target_state_id=target_state_id,
        issue_type_id=issue_type_id
    )
    
    return {
        "can_transition": can_transition,
        "source_state_id": source_state_id,
        "target_state_id": target_state_id
    }
