"""
Custom Field Services - 自定义字段业务逻辑层
"""
from typing import Optional, List, Dict, Any, Union
from datetime import date

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy.orm import selectinload

from app.models.custom_field import (
    CustomField, 
    CustomFieldOption,
    IssueCustomFieldValue,
    CustomFieldValueHistory
)
from app.models.issue import Issue
from app.models.project import Project
from app.schemas.custom_field import (
    CustomFieldCreate, 
    CustomFieldUpdate,
    CustomFieldOptionCreate,
    IssueCustomFieldValueCreate,
    IssueCustomFieldValueUpdate,
    CustomFieldTypeEnum
)
from app.core.exceptions import NotFoundException, ValidationException


# ==================== Custom Field Service ====================

async def create_custom_field(
    db: AsyncSession,
    field_data: CustomFieldCreate,
    workspace_id: int,
    user_id: int
) -> CustomField:
    """创建自定义字段"""
    # 验证项目存在（如果提供了 project_id）
    if field_data.project_id:
        project = await db.get(Project, field_data.project_id)
        if not project:
            raise NotFoundException("Project not found")
    
    # 验证数字字段范围
    if field_data.field_type == CustomFieldTypeEnum.NUMBER:
        if field_data.number_min is not None and field_data.number_max is not None:
            if field_data.number_min > field_data.number_max:
                raise ValidationException("number_min cannot be greater than number_max")
    
    # 创建字段
    field = CustomField(
        name=field_data.name,
        description=field_data.description,
        field_type=field_data.field_type.value,
        is_required=field_data.is_required,
        is_readonly=field_data.is_readonly,
        is_active=field_data.is_active,
        text_type=field_data.text_type.value if field_data.text_type else None,
        placeholder=field_data.placeholder,
        number_default=field_data.number_default,
        number_min=field_data.number_min,
        number_max=field_data.number_max,
        is_multi_select=field_data.is_multi_select,
        date_format=field_data.date_format,
        sequence=field_data.sequence,
        workspace_id=workspace_id,
        project_id=field_data.project_id,
        issue_type_id=field_data.issue_type_id,
        created_by_id=user_id
    )
    
    db.add(field)
    await db.flush()
    
    # 创建选项（如果有）
    if field_data.options and field_data.field_type == CustomFieldTypeEnum.DROPDOWN:
        for idx, opt_data in enumerate(field_data.options):
            option = CustomFieldOption(
                field_id=field.id,
                value=opt_data.value,
                color=opt_data.color,
                sequence=opt_data.sequence or idx,
                is_default=opt_data.is_default,
                is_active=opt_data.is_active,
                created_by_id=user_id
            )
            db.add(option)
    
    await db.commit()
    await db.refresh(field)
    
    # 重新加载以获取选项
    result = await db.execute(
        select(CustomField)
        .where(CustomField.id == field.id)
        .options(selectinload(CustomField.options))
    )
    return result.scalar_one()


async def get_custom_field_by_id(db: AsyncSession, field_id: int) -> CustomField:
    """获取自定义字段"""
    result = await db.execute(
        select(CustomField)
        .where(CustomField.id == field_id)
        .options(selectinload(CustomField.options))
    )
    field = result.scalar_one_or_none()
    if not field or field.is_deleted:
        raise NotFoundException("Custom field not found")
    return field


async def get_project_custom_fields(
    db: AsyncSession,
    project_id: int,
    issue_type_id: Optional[int] = None,
    include_inactive: bool = False
) -> List[CustomField]:
    """获取项目的所有自定义字段"""
    query = select(CustomField).where(
        CustomField.project_id == project_id,
        CustomField.is_deleted == False
    )
    
    if issue_type_id:
        query = query.where(
            (CustomField.issue_type_id == issue_type_id) |
            (CustomField.issue_type_id == None)  # 也获取通用字段
        )
    
    if not include_inactive:
        query = query.where(CustomField.is_active == True)
    
    query = query.order_by(CustomField.sequence)
    
    result = await db.execute(query.options(selectinload(CustomField.options)))
    return list(result.scalars().all())


async def get_workspace_custom_fields(
    db: AsyncSession,
    workspace_id: int,
    include_project_fields: bool = True
) -> List[CustomField]:
    """获取工作区的所有自定义字段（项目级和全局）"""
    query = select(CustomField).where(
        CustomField.workspace_id == workspace_id,
        CustomField.is_deleted == False,
        CustomField.is_active == True
    )
    
    if not include_project_fields:
        query = query.where(CustomField.project_id == None)
    
    query = query.order_by(CustomField.sequence)
    
    result = await db.execute(query.options(selectinload(CustomField.options)))
    return list(result.scalars().all())


async def update_custom_field(
    db: AsyncSession,
    field_id: int,
    update_data: CustomFieldUpdate
) -> CustomField:
    """更新自定义字段"""
    field = await get_custom_field_by_id(db, field_id)
    
    # 更新字段
    if update_data.name is not None:
        field.name = update_data.name
    if update_data.description is not None:
        field.description = update_data.description
    if update_data.is_required is not None:
        field.is_required = update_data.is_required
    if update_data.is_readonly is not None:
        field.is_readonly = update_data.is_readonly
    if update_data.is_active is not None:
        field.is_active = update_data.is_active
    if update_data.text_type is not None:
        field.text_type = update_data.text_type.value
    if update_data.placeholder is not None:
        field.placeholder = update_data.placeholder
    if update_data.number_default is not None:
        field.number_default = update_data.number_default
    if update_data.number_min is not None:
        field.number_min = update_data.number_min
    if update_data.number_max is not None:
        field.number_max = update_data.number_max
    if update_data.is_multi_select is not None:
        field.is_multi_select = update_data.is_multi_select
    if update_data.date_format is not None:
        field.date_format = update_data.date_format
    if update_data.sequence is not None:
        field.sequence = update_data.sequence
    
    # 验证数字范围
    if field.number_min is not None and field.number_max is not None:
        if field.number_min > field.number_max:
            raise ValidationException("number_min cannot be greater than number_max")
    
    await db.commit()
    await db.refresh(field)
    return field


async def delete_custom_field(db: AsyncSession, field_id: int):
    """删除自定义字段（软删除）"""
    field = await get_custom_field_by_id(db, field_id)
    field.is_deleted = True
    await db.commit()


# ==================== Custom Field Options Service ====================

async def create_field_option(
    db: AsyncSession,
    field_id: int,
    option_data: CustomFieldOptionCreate,
    user_id: int
) -> CustomFieldOption:
    """为下拉字段创建选项"""
    field = await get_custom_field_by_id(db, field_id)
    
    if field.field_type != CustomFieldTypeEnum.DROPDOWN.value:
        raise ValidationException("Options can only be added to dropdown fields")
    
    option = CustomFieldOption(
        field_id=field_id,
        value=option_data.value,
        color=option_data.color,
        sequence=option_data.sequence,
        is_default=option_data.is_default,
        is_active=option_data.is_active,
        created_by_id=user_id
    )
    
    db.add(option)
    await db.commit()
    await db.refresh(option)
    return option


async def update_field_option(
    db: AsyncSession,
    option_id: int,
    value: Optional[str] = None,
    color: Optional[str] = None,
    sequence: Optional[int] = None,
    is_default: Optional[bool] = None,
    is_active: Optional[bool] = None
) -> CustomFieldOption:
    """更新字段选项"""
    option = await db.get(CustomFieldOption, option_id)
    if not option:
        raise NotFoundException("Field option not found")
    
    if value is not None:
        option.value = value
    if color is not None:
        option.color = color
    if sequence is not None:
        option.sequence = sequence
    if is_default is not None:
        option.is_default = is_default
    if is_active is not None:
        option.is_active = is_active
    
    await db.commit()
    await db.refresh(option)
    return option


async def delete_field_option(db: AsyncSession, option_id: int):
    """删除字段选项"""
    option = await db.get(CustomFieldOption, option_id)
    if not option:
        raise NotFoundException("Field option not found")
    await db.delete(option)
    await db.commit()


# ==================== Issue Custom Field Value Service ====================

async def set_issue_custom_field_value(
    db: AsyncSession,
    issue_id: int,
    field_id: int,
    value_data: IssueCustomFieldValueUpdate,
    user_id: int
) -> IssueCustomFieldValue:
    """设置工作项的自定义字段值"""
    # 验证 Issue 存在
    issue = await db.get(Issue, issue_id)
    if not issue:
        raise NotFoundException("Issue not found")
    
    # 验证字段存在
    field = await get_custom_field_by_id(db, field_id)
    
    # 验证字段是否可用于此 Issue
    if field.project_id and field.project_id != issue.project_id:
        raise ValidationException("Field does not belong to this issue's project")
    
    # 验证值
    _validate_field_value(field, value_data)
    
    # 查找现有值
    result = await db.execute(
        select(IssueCustomFieldValue).where(
            IssueCustomFieldValue.issue_id == issue_id,
            IssueCustomFieldValue.field_id == field_id
        )
    )
    existing_value = result.scalar_one_or_none()
    
    # 记录旧值用于历史
    old_value_str = _serialize_value(existing_value, field.field_type) if existing_value else None
    
    if existing_value:
        # 更新现有值
        _update_value_instance(existing_value, value_data)
    else:
        # 创建新值
        value_instance = IssueCustomFieldValue(
            issue_id=issue_id,
            field_id=field_id,
            created_by_id=user_id
        )
        _update_value_instance(value_instance, value_data)
        db.add(value_instance)
        existing_value = value_instance
    
    await db.flush()
    
    # 记录历史
    new_value_str = _serialize_value(existing_value, field.field_type)
    if old_value_str != new_value_str:
        history = CustomFieldValueHistory(
            issue_id=issue_id,
            field_id=field_id,
            old_value=old_value_str,
            new_value=new_value_str,
            actor_id=user_id,
            created_by_id=user_id
        )
        db.add(history)
    
    await db.commit()
    await db.refresh(existing_value)
    return existing_value


async def get_issue_custom_field_values(
    db: AsyncSession,
    issue_id: int
) -> List[IssueCustomFieldValue]:
    """获取工作项的所有自定义字段值"""
    result = await db.execute(
        select(IssueCustomFieldValue)
        .where(IssueCustomFieldValue.issue_id == issue_id)
    )
    return list(result.scalars().all())


async def get_issue_custom_field_value(
    db: AsyncSession,
    issue_id: int,
    field_id: int
) -> Optional[IssueCustomFieldValue]:
    """获取工作项的特定自定义字段值"""
    result = await db.execute(
        select(IssueCustomFieldValue).where(
            IssueCustomFieldValue.issue_id == issue_id,
            IssueCustomFieldValue.field_id == field_id
        )
    )
    return result.scalar_one_or_none()


async def delete_issue_custom_field_value(
    db: AsyncSession,
    issue_id: int,
    field_id: int
):
    """删除工作项的自定义字段值"""
    result = await db.execute(
        select(IssueCustomFieldValue).where(
            IssueCustomFieldValue.issue_id == issue_id,
            IssueCustomFieldValue.field_id == field_id
        )
    )
    value = result.scalar_one_or_none()
    if value:
        await db.delete(value)
        await db.commit()


async def bulk_update_issue_custom_field_values(
    db: AsyncSession,
    issue_id: int,
    values: List[IssueCustomFieldValueUpdate],
    user_id: int
) -> List[IssueCustomFieldValue]:
    """批量更新工作项的自定义字段值"""
    results = []
    for value_data in values:
        result = await set_issue_custom_field_value(
            db, issue_id, value_data.field_id, value_data, user_id
        )
        results.append(result)
    return results


# ==================== Validation Functions ====================

def _validate_field_value(field: CustomField, value_data: IssueCustomFieldValueUpdate):
    """验证字段值是否符合字段定义"""
    field_type = field.field_type
    
    # 检查必填
    if field.is_required:
        has_value = (
            value_data.text_value is not None or
            value_data.number_value is not None or
            value_data.boolean_value is not None or
            value_data.date_value is not None or
            value_data.url_value is not None or
            (value_data.json_value is not None and len(value_data.json_value) > 0)
        )
        if not has_value:
            raise ValidationException(f"Field '{field.name}' is required")
    
    # 检查只读
    if field.is_readonly:
        raise ValidationException(f"Field '{field.name}' is read-only")
    
    # 数字范围验证
    if field_type == CustomFieldTypeEnum.NUMBER.value and value_data.number_value is not None:
        if field.number_min is not None and value_data.number_value < field.number_min:
            raise ValidationException(f"Value must be >= {field.number_min}")
        if field.number_max is not None and value_data.number_value > field.number_max:
            raise ValidationException(f"Value must be <= {field.number_max}")
    
    # URL 格式验证
    if field_type == CustomFieldTypeEnum.URL.value and value_data.url_value:
        if not value_data.url_value.startswith(('http://', 'https://')):
            raise ValidationException("URL must start with http:// or https://")


def _update_value_instance(instance: IssueCustomFieldValue, value_data: IssueCustomFieldValueUpdate):
    """更新值实例"""
    instance.text_value = value_data.text_value
    instance.number_value = value_data.number_value
    instance.boolean_value = value_data.boolean_value
    instance.date_value = value_data.date_value
    instance.url_value = value_data.url_value
    instance.json_value = value_data.json_value


def _serialize_value(instance: IssueCustomFieldValue, field_type: str) -> Optional[str]:
    """序列化字段值为字符串"""
    if field_type == CustomFieldTypeEnum.TEXT.value or field_type == CustomFieldTypeEnum.URL.value:
        return instance.text_value or instance.url_value
    elif field_type == CustomFieldTypeEnum.NUMBER.value:
        return str(instance.number_value) if instance.number_value is not None else None
    elif field_type == CustomFieldTypeEnum.BOOLEAN.value:
        return str(instance.boolean_value) if instance.boolean_value is not None else None
    elif field_type == CustomFieldTypeEnum.DATE.value:
        return str(instance.date_value) if instance.date_value is not None else None
    elif field_type in [CustomFieldTypeEnum.DROPDOWN.value, CustomFieldTypeEnum.MEMBER.value]:
        import json
        return json.dumps(instance.json_value) if instance.json_value else None
    return None


def get_field_value_as_display(value_instance: IssueCustomFieldValue, field: CustomField) -> Any:
    """获取字段值的显示格式"""
    field_type = field.field_type
    
    if field_type == CustomFieldTypeEnum.TEXT.value:
        return value_instance.text_value
    elif field_type == CustomFieldTypeEnum.URL.value:
        return value_instance.url_value
    elif field_type == CustomFieldTypeEnum.NUMBER.value:
        return value_instance.number_value
    elif field_type == CustomFieldTypeEnum.BOOLEAN.value:
        return value_instance.boolean_value
    elif field_type == CustomFieldTypeEnum.DATE.value:
        return value_instance.date_value
    elif field_type == CustomFieldTypeEnum.DROPDOWN.value:
        # 返回选项值而非ID
        if value_instance.json_value and field.options:
            option_map = {str(opt.id): opt.value for opt in field.options}
            return [option_map.get(v, v) for v in value_instance.json_value if v in option_map]
        return value_instance.json_value
    elif field_type == CustomFieldTypeEnum.MEMBER.value:
        # 返回成员信息
        return value_instance.json_value
    
    return None
