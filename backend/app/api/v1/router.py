from fastapi import APIRouter
from .endpoints import auth, workspace, project, ai, custom_field, project_settings, workflow, issue, cycle, module, automation, estimate_point, notification, comment, attachment

api_router = APIRouter()

api_router.include_router(auth.router, prefix="/auth", tags=["认证"])
api_router.include_router(workspace.router, prefix="/workspaces", tags=["工作空间"])
api_router.include_router(project.router, prefix="/projects", tags=["项目"])
api_router.include_router(ai.router, prefix="/ai", tags=["AI"])
api_router.include_router(custom_field.router, prefix="/custom-fields", tags=["自定义字段"])
api_router.include_router(project_settings.router, prefix="/projects/{project_id}/settings", tags=["项目设置"])
api_router.include_router(workflow.router, prefix="/projects/{project_id}/workflow", tags=["工作流与自动化"])
api_router.include_router(issue.router, prefix="/issues", tags=["工作项"])
api_router.include_router(cycle.router, prefix="/cycles", tags=["周期"])
api_router.include_router(module.router, prefix="/modules", tags=["模块"])
api_router.include_router(automation.router, prefix="/automations", tags=["自动化"])
api_router.include_router(estimate_point.router, prefix="/projects/{project_id}/estimate-points", tags=["估算点"])
api_router.include_router(notification.router, prefix="/notifications", tags=["通知"])
api_router.include_router(comment.router, prefix="/comments", tags=["评论"])
api_router.include_router(attachment.router, prefix="/attachments", tags=["附件"])