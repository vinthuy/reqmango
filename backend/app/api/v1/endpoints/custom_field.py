"""
Custom Field API Endpoints - 自定义字段管理接口
"""
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID
from typing import Optional, List

from app.db.session import get_db
from app.api.deps import get_current_user
from app.models.user import User
from app.models.project import Project
from app.schemas.custom_field import (
    CustomFieldCreate,
    CustomFieldUpdate,
    CustomFieldResponse,
    CustomFieldLite,
    CustomFieldOptionCreate,
    CustomFieldOptionUpdate,
    CustomFieldOptionResponse,
    IssueCustomFieldValueCreate,
    IssueCustomFieldValueUpdate,
    IssueCustomFieldValueResponse,
    BulkCustomFieldValueUpdate,
    IssueCustomFieldsResponse,
    CustomFieldWithValues
)
from app.services import custom_field as custom_field_service
from app.core.exceptions import NotFoundException, ValidationException

router = APIRouter()


# ==================== Custom Field CRUD ====================

@router.post("/", response_model=CustomFieldResponse, status_code=201)
async def create_custom_field(
    workspace_id: UUID,
    field_data: CustomFieldCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    创建自定义字段
    
    为工作空间或项目创建一个新的自定义字段。
    支持的字段类型：text, number, dropdown, boolean, date, member, url
    
    下拉类型字段可以同时创建选项。
    """
    # 验证项目访问权限（如果提供了 project_id）
    if field_data.project_id:
        project = await db.get(Project, field_data.project_id)
        if not project or project.is_deleted:
            raise HTTPException(status_code=404, detail="Project not found")
    
    field = await custom_field_service.create_custom_field(
        db=db,
        field_data=field_data,
        workspace_id=workspace_id,
        user_id=current_user.id
    )
    return field


@router.get("/", response_model=List[CustomFieldResponse])
async def list_custom_fields(
    workspace_id: UUID,
    project_id: Optional[UUID] = None,
    issue_type_id: Optional[UUID] = None,
    include_inactive: bool = False,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    列出自定义字段
    
    - 如果提供 project_id，列出该项目的工作项类型字段
    - 如果不提供 project_id，列出工作空间级别的字段
    - 可以按 issue_type_id 筛选
    """
    if project_id:
        fields = await custom_field_service.get_project_custom_fields(
            db=db,
            project_id=project_id,
            issue_type_id=issue_type_id,
            include_inactive=include_inactive
        )
    else:
        fields = await custom_field_service.get_workspace_custom_fields(
            db=db,
            workspace_id=workspace_id,
            include_project_fields=True
        )
    
    return fields


@router.get("/{field_id}", response_model=CustomFieldResponse)
async def get_custom_field(
    field_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取自定义字段详情
    """
    field = await custom_field_service.get_custom_field_by_id(db, field_id)
    return field


@router.put("/{field_id}", response_model=CustomFieldResponse)
async def update_custom_field(
    field_id: UUID,
    update_data: CustomFieldUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新自定义字段
    """
    field = await custom_field_service.update_custom_field(
        db=db,
        field_id=field_id,
        update_data=update_data
    )
    return field


@router.delete("/{field_id}", status_code=204)
async def delete_custom_field(
    field_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除自定义字段（软删除）
    """
    await custom_field_service.delete_custom_field(db, field_id)
    return None


# ==================== Custom Field Options ====================

@router.post("/{field_id}/options", response_model=CustomFieldOptionResponse, status_code=201)
async def create_field_option(
    field_id: UUID,
    option_data: CustomFieldOptionCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    为下拉字段创建选项
    """
    option = await custom_field_service.create_field_option(
        db=db,
        field_id=field_id,
        option_data=option_data,
        user_id=current_user.id
    )
    return option


@router.put("/{field_id}/options/{option_id}", response_model=CustomFieldOptionResponse)
async def update_field_option(
    field_id: UUID,
    option_id: UUID,
    value: Optional[str] = None,
    color: Optional[str] = None,
    sequence: Optional[int] = None,
    is_default: Optional[bool] = None,
    is_active: Optional[bool] = None,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新字段选项
    """
    option = await custom_field_service.update_field_option(
        db=db,
        option_id=option_id,
        value=value,
        color=color,
        sequence=sequence,
        is_default=is_default,
        is_active=is_active
    )
    return option


@router.delete("/{field_id}/options/{option_id}", status_code=204)
async def delete_field_option(
    field_id: UUID,
    option_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除字段选项
    """
    await custom_field_service.delete_field_option(db, option_id)
    return None


# ==================== Issue Custom Field Values ====================

@router.post("/issues/{issue_id}/values", response_model=IssueCustomFieldValueResponse, status_code=201)
async def set_issue_custom_field_value(
    issue_id: UUID,
    value_data: IssueCustomFieldValueCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    设置工作项的自定义字段值
    
    可以设置文本、数字、布尔、日期、URL 或下拉类型的值。
    下拉和成员类型需要传入选项ID列表（json_value）。
    """
    update_data = IssueCustomFieldValueUpdate(
        text_value=value_data.text_value,
        number_value=value_data.number_value,
        boolean_value=value_data.boolean_value,
        date_value=value_data.date_value,
        url_value=value_data.url_value,
        json_value=value_data.json_value
    )
    
    value = await custom_field_service.set_issue_custom_field_value(
        db=db,
        issue_id=issue_id,
        field_id=value_data.field_id,
        value_data=update_data,
        user_id=current_user.id
    )
    
    # 获取字段信息
    field = await custom_field_service.get_custom_field_by_id(db, value_data.field_id)
    
    return IssueCustomFieldValueResponse(
        id=value.id,
        issue_id=value.issue_id,
        field_id=value.field_id,
        field_name=field.name,
        field_type=field.field_type,
        value=custom_field_service.get_field_value_as_display(value, field),
        text_value=value.text_value,
        number_value=value.number_value,
        boolean_value=value.boolean_value,
        date_value=value.date_value,
        url_value=value.url_value,
        json_value=value.json_value
    )


@router.get("/issues/{issue_id}/values", response_model=List[IssueCustomFieldValueResponse])
async def list_issue_custom_field_values(
    issue_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作项的所有自定义字段值
    """
    values = await custom_field_service.get_issue_custom_field_values(db, issue_id)
    
    results = []
    for value in values:
        field = await custom_field_service.get_custom_field_by_id(db, value.field_id)
        results.append(IssueCustomFieldValueResponse(
            id=value.id,
            issue_id=value.issue_id,
            field_id=value.field_id,
            field_name=field.name,
            field_type=field.field_type,
            value=custom_field_service.get_field_value_as_display(value, field),
            text_value=value.text_value,
            number_value=value.number_value,
            boolean_value=value.boolean_value,
            date_value=value.date_value,
            url_value=value.url_value,
            json_value=value.json_value
        ))
    
    return results


@router.put("/issues/{issue_id}/values/{field_id}", response_model=IssueCustomFieldValueResponse)
async def update_issue_custom_field_value(
    issue_id: UUID,
    field_id: UUID,
    value_data: IssueCustomFieldValueUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    更新工作项的特定自定义字段值
    """
    value = await custom_field_service.set_issue_custom_field_value(
        db=db,
        issue_id=issue_id,
        field_id=field_id,
        value_data=value_data,
        user_id=current_user.id
    )
    
    # 获取字段信息
    field = await custom_field_service.get_custom_field_by_id(db, field_id)
    
    return IssueCustomFieldValueResponse(
        id=value.id,
        issue_id=value.issue_id,
        field_id=value.field_id,
        field_name=field.name,
        field_type=field.field_type,
        value=custom_field_service.get_field_value_as_display(value, field),
        text_value=value.text_value,
        number_value=value.number_value,
        boolean_value=value.boolean_value,
        date_value=value.date_value,
        url_value=value.url_value,
        json_value=value.json_value
    )


@router.delete("/issues/{issue_id}/values/{field_id}", status_code=204)
async def delete_issue_custom_field_value(
    issue_id: UUID,
    field_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    删除工作项的自定义字段值
    """
    await custom_field_service.delete_issue_custom_field_value(db, issue_id, field_id)
    return None


@router.post("/issues/{issue_id}/values/bulk", response_model=List[IssueCustomFieldValueResponse])
async def bulk_update_issue_custom_field_values(
    issue_id: UUID,
    bulk_data: BulkCustomFieldValueUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    批量更新工作项的自定义字段值
    """
    values = await custom_field_service.bulk_update_issue_custom_field_values(
        db=db,
        issue_id=issue_id,
        values=bulk_data.values,
        user_id=current_user.id
    )
    
    results = []
    for value in values:
        field = await custom_field_service.get_custom_field_by_id(db, value.field_id)
        results.append(IssueCustomFieldValueResponse(
            id=value.id,
            issue_id=value.issue_id,
            field_id=value.field_id,
            field_name=field.name,
            field_type=field.field_type,
            value=custom_field_service.get_field_value_as_display(value, field),
            text_value=value.text_value,
            number_value=value.number_value,
            boolean_value=value.boolean_value,
            date_value=value.date_value,
            url_value=value.url_value,
            json_value=value.json_value
        ))
    
    return results


# ==================== Issue Custom Fields with Definitions ====================

@router.get("/issues/{issue_id}/fields", response_model=IssueCustomFieldsResponse)
async def get_issue_custom_fields_with_definitions(
    issue_id: UUID,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """
    获取工作项的所有自定义字段定义及其值
    
    返回字段定义和对应的值，便于前端渲染表单。
    """
    from app.models.issue import Issue
    from app.services import custom_field as cf_service
    
    # 获取 Issue
    issue = await db.get(Issue, issue_id)
    if not issue:
        raise HTTPException(status_code=404, detail="Issue not found")
    
    # 获取项目的自定义字段
    fields = await cf_service.get_project_custom_fields(
        db=db,
        project_id=issue.project_id,
        include_inactive=False
    )
    
    # 获取工作项的字段值
    values = await cf_service.get_issue_custom_field_values(db, issue_id)
    values_map = {str(v.field_id): v for v in values}
    
    # 组装结果
    fields_with_values = []
    for field in fields:
        field_lite = CustomFieldLite(
            id=field.id,
            name=field.name,
            field_type=field.field_type,
            is_required=field.is_required,
            is_readonly=field.is_readonly,
            options=[
                CustomFieldOptionResponse(
                    id=opt.id,
                    value=opt.value,
                    color=opt.color,
                    sequence=opt.sequence,
                    is_default=opt.is_default,
                    is_active=opt.is_active
                )
                for opt in field.options if opt.is_active
            ]
        )
        
        value_instance = values_map.get(str(field.id))
        value = cf_service.get_field_value_as_display(value_instance, field) if value_instance else None
        
        fields_with_values.append(CustomFieldWithValues(
            field=field_lite,
            value=value
        ))
    
    return IssueCustomFieldsResponse(
        issue_id=issue_id,
        fields=fields_with_values
    )
