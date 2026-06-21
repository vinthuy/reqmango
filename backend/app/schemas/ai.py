from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any, Literal
from datetime import datetime, date
from enum import Enum

class AIMode(str, Enum):
    ASK = "ask"
    BUILD = "build"

class AIIntent(str, Enum):
    SEARCH = "search"
    CREATE = "create"
    UPDATE = "update"
    ANALYZE = "analyze"
    HELP = "help"

class AIMessage(BaseModel):
    content: str = Field(..., min_length=1)
    mode: AIMode = AIMode.ASK
    context: Optional[Dict[str, Any]] = None
    attachments: Optional[List[str]] = []

class AIRequest(BaseModel):
    message: AIMessage
    workspace_id: Optional[int] = None
    project_id: Optional[int] = None
    thread_id: Optional[int] = None

class AIAction(BaseModel):
    action_type: str
    target_type: str
    target_id: Optional[int] = None
    changes: Dict[str, Any] = {}
    description: str

class AIPlan(BaseModel):
    actions: List[AIAction]
    requires_confirmation: bool = True
    estimated_impact: str

class AIResponse(BaseModel):
    content: str
    intent: AIIntent
    plan: Optional[AIPlan] = None
    results: Optional[List[Dict[str, Any]]] = None
    suggestions: Optional[List[str]] = None
    thread_id: int

class AIThread(BaseModel):
    id: int
    title: str
    messages: List[Dict[str, Any]]
    created_at: datetime
    updated_at: datetime


class PriorityLevel(str, Enum):
    """优先级级别"""
    URGENT = "urgent"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"


class TaskExtractionRequest(BaseModel):
    """任务提取请求"""
    text: str = Field(..., min_length=1, description="自然语言输入文本")
    workspace_id: int = Field(..., description="工作区ID")
    project_id: int = Field(..., description="项目ID")
    auto_create: bool = Field(False, description="是否自动创建任务")


class TaskExtractionResponse(BaseModel):
    """任务提取响应"""
    extracted_data: Dict[str, Any] = Field(..., description="提取的数据")
    confidence: float = Field(..., ge=0.0, le=1.0, description="置信度")
    warnings: List[str] = Field(default_factory=list, description="警告信息")
    parsing_notes: List[str] = Field(default_factory=list, description="解析说明")
    suggestions: List[str] = Field(default_factory=list, description="建议")


class TaskCreateFromNLPRequest(BaseModel):
    """从NLP创建任务请求"""
    text: str = Field(..., min_length=1, description="自然语言输入文本")
    workspace_id: int = Field(..., description="工作区ID")
    project_id: int = Field(..., description="项目ID")
    auto_create: bool = Field(True, description="是否自动创建任务")
    review_before_create: bool = Field(False, description="创建前是否需要审核")


class TaskCreateFromNLPResponse(BaseModel):
    """从NLP创建任务响应"""
    task_id: Optional[int] = None
    created: bool = Field(..., description="是否已创建")
    extraction_details: Dict[str, Any] = Field(..., description="提取详情")
    confidence: float = Field(..., description="置信度")
    warnings: List[str] = Field(default_factory=list, description="警告信息")

