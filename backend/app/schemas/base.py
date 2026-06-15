from datetime import datetime
from uuid import UUID, uuid4
from pydantic import BaseModel, ConfigDict, Field
from typing import Optional, Any

class BaseSchema(BaseModel):
    model_config = ConfigDict(
        from_attributes=True,
        populate_by_name=True,
        use_enum_values=True,
    )

class TimestampSchema(BaseSchema):
    created_at: datetime
    updated_at: datetime

class UUIDSchema(BaseSchema):
    id: UUID = Field(default_factory=uuid4)

class AuditSchema(UUIDSchema, TimestampSchema):
    created_by: Optional[UUID] = None
    updated_by: Optional[UUID] = None

class SoftDeleteSchema(BaseSchema):
    deleted_at: Optional[datetime] = None
    is_deleted: bool = False

class PaginatedResponse(BaseSchema):
    results: list[Any]
    total: int
    page: int
    per_page: int
    has_next: bool
    has_prev: bool