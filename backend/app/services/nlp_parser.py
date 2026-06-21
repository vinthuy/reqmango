"""
NLP Parser Service - 自然语言解析器
实现从自然语言文本中提取任务信息的功能
支持中文和英文输入
"""
from typing import Optional, List, Dict, Any
from datetime import datetime, date, timedelta
import re
from enum import Enum

from pydantic import BaseModel, Field


class PriorityLevel(str, Enum):
    URGENT = "urgent"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"


class TaskExtraction(BaseModel):
    """任务提取结果"""
    title: Optional[str] = None
    description: Optional[str] = None
    priority: Optional[PriorityLevel] = None
    due_date: Optional[date] = None
    start_date: Optional[date] = None
    assignee_email: Optional[str] = None
    assignee_name: Optional[str] = None
    labels: List[str] = []
    confidence: float = 0.0
    raw_input: str = ""
    parsing_notes: List[str] = []


class DateExpression:
    """日期表达式解析器"""
    
    # 中文日期关键词映射
    CHINESE_DATE_PATTERNS = {
        "今天": 0,
        "明天": 1,
        "后天": 2,
        "大后天": 3,
        "本周末": None,  # 特殊处理
        "下周一": None,
        "下周二": None,
        "下周三": None,
        "下周四": None,
        "下周五": None,
        "下周": None,
        "本月末": None,
        "下月末": None,
    }
    
    @staticmethod
    def parse_relative_date(text: str) -> Optional[date]:
        """解析相对日期表达式"""
        today = date.today()
        
        # 检查简单偏移量（今天、明天、后天）
        for keyword, offset in DateExpression.CHINESE_DATE_PATTERNS.items():
            if keyword in text and offset is not None:
                return today + timedelta(days=offset)
        
        # 处理"X天后"格式
        match = re.search(r'(\d+)\s*天后', text)
        if match:
            days = int(match.group(1))
            return today + timedelta(days=days)
        
        # 处理"X周后"格式
        match = re.search(r'(\d+)\s*周后', text)
        if match:
            weeks = int(match.group(1))
            return today + timedelta(weeks=weeks)
        
        # 处理英文相对日期
        english_patterns = {
            "today": 0,
            "tomorrow": 1,
            "next week": 7,
            "in a week": 7,
            "in two weeks": 14,
            "in a month": 30,
        }
        
        text_lower = text.lower()
        for pattern, offset in english_patterns.items():
            if pattern in text_lower:
                return today + timedelta(days=offset)
        
        return None
    
    @staticmethod
    def parse_absolute_date(text: str) -> Optional[date]:
        """解析绝对日期"""
        # 匹配 YYYY-MM-DD 格式
        match = re.search(r'(\d{4})-(\d{1,2})-(\d{1,2})', text)
        if match:
            try:
                year, month, day = int(match.group(1)), int(match.group(2)), int(match.group(3))
                return date(year, month, day)
            except ValueError:
                pass
        
        # 匹配 MM/DD/YYYY 格式
        match = re.search(r'(\d{1,2})/(\d{1,2})/(\d{4})', text)
        if match:
            try:
                month, day, year = int(match.group(1)), int(match.group(2)), int(match.group(3))
                return date(year, month, day)
            except ValueError:
                pass
        
        # 匹配中文日期格式：2024年1月15日
        match = re.search(r'(\d{4})年(\d{1,2})月(\d{1,2})日', text)
        if match:
            try:
                year, month, day = int(match.group(1)), int(match.group(2)), int(match.group(3))
                return date(year, month, day)
            except ValueError:
                pass
        
        return None
    
    @staticmethod
    def parse_weekday(text: str) -> Optional[date]:
        """解析星期几"""
        today = date.today()
        weekday_map = {
            "周一": 0, "星期一": 0, "monday": 0,
            "周二": 1, "星期二": 1, "tuesday": 1,
            "周三": 2, "星期三": 2, "wednesday": 2,
            "周四": 3, "星期四": 3, "thursday": 3,
            "周五": 4, "星期五": 4, "friday": 4,
            "周六": 5, "星期六": 5, "saturday": 5,
            "周日": 6, "星期日": 6, "sunday": 6,
        }
        
        text_lower = text.lower()
        for keyword, weekday in weekday_map.items():
            if keyword in text_lower:
                # 计算下一个该星期的日期
                days_ahead = weekday - today.weekday()
                if days_ahead <= 0:  # 如果已经过了，就下周
                    days_ahead += 7
                return today + timedelta(days=days_ahead)
        
        return None


class PriorityExtractor:
    """优先级提取器"""
    
    PRIORITY_KEYWORDS = {
        PriorityLevel.URGENT: [
            "urgent", "critical", "blocker", "asap", "紧急", "严重", "立刻", "马上", "火烧眉毛"
        ],
        PriorityLevel.HIGH: [
            "high", "important", "priority", "高", "重要", "优先"
        ],
        PriorityLevel.MEDIUM: [
            "medium", "normal", "中", "普通", "一般"
        ],
        PriorityLevel.LOW: [
            "low", "minor", "nice to have", "低", "次要", "有空再做"
        ],
    }
    
    @staticmethod
    def extract(text: str) -> Optional[PriorityLevel]:
        """从文本中提取优先级"""
        text_lower = text.lower()
        
        for priority, keywords in PriorityExtractor.PRIORITY_KEYWORDS.items():
            if any(keyword in text_lower for keyword in keywords):
                return priority
        
        return None


class AssigneeExtractor:
    """负责人提取器"""
    
    # 常见的人名模式（简化版）
    NAME_PATTERNS = [
        r'分配给\s*([\u4e00-\u9fa5]{2,4})',  # 分配给张三
        r'负责人[:：]\s*([\u4e00-\u9fa5]{2,4})',  # 负责人：张三
        r'assign(?:ed)?\s*(?:to)?\s*([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)',  # assigned to John Doe
        r'by\s+([A-Z][a-z]+)',  # by John
    ]
    
    EMAIL_PATTERN = r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    
    @staticmethod
    def extract(text: str) -> Dict[str, Optional[str]]:
        """提取负责人信息"""
        result = {"name": None, "email": None}
        
        # 尝试提取邮箱
        email_match = re.search(AssigneeExtractor.EMAIL_PATTERN, text)
        if email_match:
            result["email"] = email_match.group(0)
        
        # 尝试提取中文名
        for pattern in AssigneeExtractor.NAME_PATTERNS:
            match = re.search(pattern, text)
            if match:
                name = match.group(1).strip()
                # 判断是中文还是英文名
                if re.match(r'^[\u4e00-\u9fa5]+$', name):
                    result["name"] = name
                else:
                    result["name"] = name
                break
        
        return result


class LabelExtractor:
    """标签提取器"""
    
    LABEL_PATTERNS = [
        r'标签[:：]\s*([^\n,;，；]+)',  # 标签：bug, frontend
        r'labels?[:：]\s*([^\n,;，；]+)',  # labels: bug
        r'#(\w+)',  # #bug #frontend
    ]
    
    @staticmethod
    def extract(text: str) -> List[str]:
        """提取标签列表"""
        labels = set()
        
        # 提取 #tag 格式的标签
        hashtag_matches = re.findall(r'#(\w+)', text)
        labels.update(hashtag_matches)
        
        # 提取 "标签："格式的标签
        for pattern in LabelExtractor.LABEL_PATTERNS[:2]:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                label_text = match.group(1)
                # 分割多个标签
                split_labels = re.split(r'[,;，；\s]+', label_text)
                labels.update([l.strip() for l in split_labels if l.strip()])
        
        return list(labels)


class TitleExtractor:
    """标题提取器"""
    
    TITLE_INDICATORS = [
        "标题", "title", "主题", "subject",
        "名为", "called", "named",
    ]
    
    @staticmethod
    def extract(text: str) -> Optional[str]:
        """提取任务标题"""
        # 尝试找到明确的标题指示词
        for indicator in TitleExtractor.TITLE_INDICATORS:
            if indicator in text.lower():
                # 提取指示词后的内容直到下一个关键字
                pattern = rf'{indicator}[:：\s]*([^,;，；。\.]+)'
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    return match.group(1).strip()
        
        # 如果没有明确指示，尝试提取第一个有意义的短语
        # 移除常见的任务创建前缀
        cleaned = re.sub(
            r'^(创建|新建|add|create)\s*(一个|an|a)?\s*(任务|task|issue|bug)?\s*',
            '',
            text,
            flags=re.IGNORECASE
        )
        
        # 取第一句话作为标题
        sentences = re.split(r'[。\.!?！？]', cleaned)
        if sentences and sentences[0].strip():
            title = sentences[0].strip()
            # 限制标题长度
            if len(title) > 100:
                title = title[:100] + "..."
            return title
        
        return None


class DescriptionExtractor:
    """描述提取器"""
    
    DESCRIPTION_INDICATORS = [
        "描述", "description", "详情", "details",
        "内容是", "内容为", "内容是",
    ]
    
    @staticmethod
    def extract(text: str) -> Optional[str]:
        """提取任务描述"""
        # 尝试找到明确的描述指示词
        for indicator in DescriptionExtractor.DESCRIPTION_INDICATORS:
            if indicator in text.lower():
                # 提取指示词后的所有内容
                pattern = rf'{indicator}[:：\s]*(.+)'
                match = re.search(pattern, text, re.IGNORECASE | re.DOTALL)
                if match:
                    return match.group(1).strip()
        
        return None


class NLPParser:
    """自然语言解析器主类"""
    
    def __init__(self):
        self.priority_extractor = PriorityExtractor()
        self.assignee_extractor = AssigneeExtractor()
        self.label_extractor = LabelExtractor()
        self.title_extractor = TitleExtractor()
        self.description_extractor = DescriptionExtractor()
        self.date_parser = DateExpression()
    
    def parse_task_creation(self, text: str) -> TaskExtraction:
        """
        解析任务创建的自然语言输入
        
        Args:
            text: 用户输入的自然语言文本
            
        Returns:
            TaskExtraction: 提取的任务信息
        """
        extraction = TaskExtraction(raw_input=text)
        notes = []
        
        # 1. 提取标题
        title = self.title_extractor.extract(text)
        if title:
            extraction.title = title
            notes.append("Title extracted")
        
        # 2. 提取描述
        description = self.description_extractor.extract(text)
        if description:
            extraction.description = description
            notes.append("Description extracted")
        
        # 3. 提取优先级
        priority = self.priority_extractor.extract(text)
        if priority:
            extraction.priority = priority
            notes.append(f"Priority detected: {priority.value}")
        
        # 4. 提取截止日期
        due_date = self._extract_due_date(text)
        if due_date:
            extraction.due_date = due_date
            notes.append(f"Due date detected: {due_date}")
        
        # 5. 提取开始日期
        start_date = self._extract_start_date(text)
        if start_date:
            extraction.start_date = start_date
            notes.append(f"Start date detected: {start_date}")
        
        # 6. 提取负责人
        assignee_info = self.assignee_extractor.extract(text)
        if assignee_info["email"]:
            extraction.assignee_email = assignee_info["email"]
            notes.append(f"Assignee email detected: {assignee_info['email']}")
        if assignee_info["name"]:
            extraction.assignee_name = assignee_info["name"]
            notes.append(f"Assignee name detected: {assignee_info['name']}")
        
        # 7. 提取标签
        labels = self.label_extractor.extract(text)
        if labels:
            extraction.labels = labels
            notes.append(f"Labels detected: {', '.join(labels)}")
        
        # 8. 计算置信度
        extraction.confidence = self._calculate_confidence(extraction)
        extraction.parsing_notes = notes
        
        return extraction
    
    def _extract_due_date(self, text: str) -> Optional[date]:
        """提取截止日期"""
        # 查找截止日期相关关键词
        due_keywords = ["截止", "due", "deadline", "到期", "完成日期"]
        has_due_keyword = any(kw in text.lower() for kw in due_keywords)
        
        if has_due_keyword:
            # 尝试解析绝对日期
            absolute_date = self.date_parser.parse_absolute_date(text)
            if absolute_date:
                return absolute_date
            
            # 尝试解析相对日期
            relative_date = self.date_parser.parse_relative_date(text)
            if relative_date:
                return relative_date
            
            # 尝试解析星期几
            weekday_date = self.date_parser.parse_weekday(text)
            if weekday_date:
                return weekday_date
        
        return None
    
    def _extract_start_date(self, text: str) -> Optional[date]:
        """提取开始日期"""
        start_keywords = ["开始", "start", "begin", "启动"]
        has_start_keyword = any(kw in text.lower() for kw in start_keywords)
        
        if has_start_keyword:
            # 类似截止日期的逻辑
            absolute_date = self.date_parser.parse_absolute_date(text)
            if absolute_date:
                return absolute_date
            
            relative_date = self.date_parser.parse_relative_date(text)
            if relative_date:
                return relative_date
        
        return None
    
    def _calculate_confidence(self, extraction: TaskExtraction) -> float:
        """
        计算解析置信度
        
        基于提取到的字段数量和质量计算置信度
        """
        score = 0.0
        max_score = 0.0
        
        # 标题（权重最高）
        max_score += 3.0
        if extraction.title:
            score += 3.0
            # 标题质量检查
            if len(extraction.title) > 5:
                score += 0.5
        
        # 优先级
        max_score += 2.0
        if extraction.priority:
            score += 2.0
        
        # 截止日期
        max_score += 2.0
        if extraction.due_date:
            score += 2.0
        
        # 负责人
        max_score += 1.5
        if extraction.assignee_email or extraction.assignee_name:
            score += 1.5
        
        # 标签
        max_score += 1.0
        if extraction.labels:
            score += 1.0
        
        # 描述
        max_score += 0.5
        if extraction.description:
            score += 0.5
        
        # 归一化到 0-1 范围
        confidence = score / max_score if max_score > 0 else 0.0
        
        return round(min(confidence, 1.0), 2)
    
    def validate_extraction(self, extraction: TaskExtraction) -> List[str]:
        """
        验证提取结果的完整性
        
        Returns:
            缺失字段的警告列表
        """
        warnings = []
        
        if not extraction.title:
            warnings.append("Title is missing - task may be unclear")
        
        if not extraction.priority:
            warnings.append("Priority not detected - default will be used")
        
        if not extraction.due_date:
            warnings.append("Due date not specified")
        
        if not extraction.assignee_email and not extraction.assignee_name:
            warnings.append("No assignee detected - task will be unassigned")
        
        if extraction.confidence < 0.5:
            warnings.append("Low confidence - please review extracted information")
        
        return warnings


# 使用示例
if __name__ == "__main__":
    parser = NLPParser()
    
    # 测试用例1：中文输入
    test_text_1 = "创建一个高优先级任务，标题为修复登录页面bug，描述为用户反馈使用Chrome浏览器时无法登录，截止日期为本周五，负责人分配给张三，标签：bug, frontend"
    result_1 = parser.parse_task_creation(test_text_1)
    print("Test 1 - Chinese Input:")
    print(f"  Title: {result_1.title}")
    print(f"  Priority: {result_1.priority}")
    print(f"  Due Date: {result_1.due_date}")
    print(f"  Assignee: {result_1.assignee_name}")
    print(f"  Labels: {result_1.labels}")
    print(f"  Confidence: {result_1.confidence}")
    print(f"  Notes: {result_1.parsing_notes}")
    print()
    
    # 测试用例2：英文输入
    test_text_2 = "Create urgent task: Fix login page bug. User cannot login with Chrome browser. Due tomorrow, assign to john@example.com #bug #frontend"
    result_2 = parser.parse_task_creation(test_text_2)
    print("Test 2 - English Input:")
    print(f"  Title: {result_2.title}")
    print(f"  Priority: {result_2.priority}")
    print(f"  Due Date: {result_2.due_date}")
    print(f"  Assignee Email: {result_2.assignee_email}")
    print(f"  Labels: {result_2.labels}")
    print(f"  Confidence: {result_2.confidence}")
    print(f"  Notes: {result_2.parsing_notes}")
