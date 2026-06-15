from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any, Literal
from uuid import UUID
from datetime import datetime
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
    workspace_id: Optional[UUID] = None
    project_id: Optional[UUID] = None
    thread_id: Optional[UUID] = None

class AIAction(BaseModel):
    action_type: str
    target_type: str
    target_id: Optional[UUID] = None
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
    thread_id: UUID

class AIThread(BaseModel):
    id: UUID
    title: str
    messages: List[Dict[str, Any]]
    created_at: datetime
    updated_at: datetime