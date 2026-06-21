"""
NLP Parser Tests - 测试自然语言解析器功能
"""
import pytest
from datetime import date, timedelta
from app.services.nlp_parser import (
    NLPParser, TaskExtraction, PriorityLevel,
    PriorityExtractor, AssigneeExtractor, LabelExtractor,
    TitleExtractor, DescriptionExtractor, DateExpression
)


class TestPriorityExtractor:
    """优先级提取器测试"""
    
    def test_extract_urgent_priority_chinese(self):
        """测试中文紧急优先级"""
        assert PriorityExtractor.extract("这是一个紧急任务") == PriorityLevel.URGENT
        assert PriorityExtractor.extract("严重问题需要立即处理") == PriorityLevel.URGENT
        assert PriorityExtractor.extract("火烧眉毛的事情") == PriorityLevel.URGENT
    
    def test_extract_urgent_priority_english(self):
        """测试英文紧急优先级"""
        assert PriorityExtractor.extract("This is urgent") == PriorityLevel.URGENT
        assert PriorityExtractor.extract("Critical bug fix needed") == PriorityLevel.URGENT
        assert PriorityExtractor.extract("ASAP please") == PriorityLevel.URGENT
    
    def test_extract_high_priority(self):
        """测试高优先级"""
        assert PriorityExtractor.extract("高优先级任务") == PriorityLevel.HIGH
        assert PriorityExtractor.extract("Important task") == PriorityLevel.HIGH
    
    def test_extract_medium_priority(self):
        """测试中优先级"""
        assert PriorityExtractor.extract("中等优先级") == PriorityLevel.MEDIUM
        assert PriorityExtractor.extract("Normal task") == PriorityLevel.MEDIUM
    
    def test_extract_low_priority(self):
        """测试低优先级"""
        assert PriorityExtractor.extract("低优先级") == PriorityLevel.LOW
        assert PriorityExtractor.extract("Low priority task") == PriorityLevel.LOW
    
    def test_no_priority_detected(self):
        """测试无优先级关键词"""
        assert PriorityExtractor.extract("Just a regular task") is None


class TestAssigneeExtractor:
    """负责人提取器测试"""
    
    def test_extract_chinese_name(self):
        """测试提取中文姓名"""
        result = AssigneeExtractor.extract("分配给张三")
        assert result["name"] == "张三"
        assert result["email"] is None
    
    def test_extract_email(self):
        """测试提取邮箱"""
        result = AssigneeExtractor.extract("assign to john@example.com")
        assert result["email"] == "john@example.com"
    
    def test_extract_assignee_with_colon(self):
        """测试冒号格式的负责人"""
        result = AssigneeExtractor.extract("负责人：李四")
        assert result["name"] == "李四"
    
    def test_extract_english_name(self):
        """测试提取英文姓名"""
        result = AssigneeExtractor.extract("assigned to John Doe")
        assert result["name"] == "John Doe"
    
    def test_no_assignee(self):
        """测试无负责人信息"""
        result = AssigneeExtractor.extract("Create a task")
        assert result["name"] is None
        assert result["email"] is None


class TestLabelExtractor:
    """标签提取器测试"""
    
    def test_extract_hashtag_labels(self):
        """测试#标签格式"""
        labels = LabelExtractor.extract("Task #bug #frontend #urgent")
        assert "bug" in labels
        assert "frontend" in labels
        assert "urgent" in labels
    
    def test_extract_label_prefix_format(self):
        """测试标签：格式"""
        labels = LabelExtractor.extract("标签：bug, frontend, backend")
        assert "bug" in labels
        assert "frontend" in labels
        assert "backend" in labels
    
    def test_extract_labels_semicolon_separator(self):
        """测试分号分隔的标签"""
        labels = LabelExtractor.extract("labels: bug; feature; enhancement")
        assert "bug" in labels
        assert "feature" in labels
    
    def test_no_labels(self):
        """测试无标签"""
        labels = LabelExtractor.extract("Just a task without labels")
        assert len(labels) == 0


class TestTitleExtractor:
    """标题提取器测试"""
    
    def test_extract_title_with_indicator(self):
        """测试带指示词的标题"""
        title = TitleExtractor.extract("创建一个任务，标题为修复登录bug")
        assert title is not None
        assert "修复登录bug" in title
    
    def test_extract_title_english(self):
        """测试英文标题"""
        title = TitleExtractor.extract("Create task: Fix login page")
        assert title is not None
        assert "Fix login page" in title
    
    def test_extract_title_from_sentence(self):
        """测试从句子中提取标题"""
        title = TitleExtractor.extract("创建一个修复登录页面bug的任务")
        assert title is not None
    
    def test_title_length_limit(self):
        """测试标题长度限制"""
        long_text = "创建" + "a" * 200 + "的任务"
        title = TitleExtractor.extract(long_text)
        assert len(title) <= 103  # 100 + "..."


class TestDescriptionExtractor:
    """描述提取器测试"""
    
    def test_extract_description_with_indicator(self):
        """测试带指示词的描述"""
        desc = DescriptionExtractor.extract("任务描述为用户反馈无法登录")
        assert desc is not None
        assert "用户反馈无法登录" in desc
    
    def test_extract_description_english(self):
        """测试英文描述"""
        desc = DescriptionExtractor.extract("Description: User cannot login with Chrome")
        assert desc is not None
        assert "User cannot login with Chrome" in desc
    
    def test_no_description(self):
        """测试无描述"""
        desc = DescriptionExtractor.extract("Create a task")
        assert desc is None


class TestDateExpression:
    """日期表达式解析器测试"""
    
    def test_parse_today(self):
        """测试今天"""
        today = date.today()
        result = DateExpression.parse_relative_date("今天完成任务")
        assert result == today
    
    def test_parse_tomorrow(self):
        """测试明天"""
        tomorrow = date.today() + timedelta(days=1)
        result = DateExpression.parse_relative_date("明天完成")
        assert result == tomorrow
    
    def test_parse_days_later(self):
        """测试X天后"""
        expected = date.today() + timedelta(days=5)
        result = DateExpression.parse_relative_date("5天后完成")
        assert result == expected
    
    def test_parse_absolute_date_yyyy_mm_dd(self):
        """测试YYYY-MM-DD格式"""
        result = DateExpression.parse_absolute_date("截止日期2024-12-25")
        assert result == date(2024, 12, 25)
    
    def test_parse_chinese_date(self):
        """测试中文日期格式"""
        result = DateExpression.parse_absolute_date("2024年1月15日截止")
        assert result == date(2024, 1, 15)
    
    def test_parse_weekday(self):
        """测试星期几"""
        result = DateExpression.parse_weekday("下周一完成")
        assert result is not None
        # 验证是星期一
        assert result.weekday() == 0


class TestNLPParser:
    """NLP解析器主类测试"""
    
    def setup_method(self):
        """每个测试前初始化解析器"""
        self.parser = NLPParser()
    
    def test_parse_chinese_task_creation(self):
        """测试中文任务创建"""
        text = "创建一个高优先级任务，标题为修复登录页面bug，描述为用户反馈使用Chrome浏览器时无法登录，截止日期为本周五，负责人分配给张三，标签：bug, frontend"
        result = self.parser.parse_task_creation(text)
        
        assert result.title is not None
        assert result.priority == PriorityLevel.HIGH
        assert result.assignee_name == "张三"
        assert "bug" in result.labels
        assert "frontend" in result.labels
        assert result.confidence > 0.7
    
    def test_parse_english_task_creation(self):
        """测试英文任务创建"""
        text = "Create urgent task: Fix login page bug. User cannot login with Chrome browser. Due tomorrow, assign to john@example.com #bug #frontend"
        result = self.parser.parse_task_creation(text)
        
        assert result.title is not None
        assert result.priority == PriorityLevel.URGENT
        assert result.assignee_email == "john@example.com"
        assert "bug" in result.labels
        assert result.confidence > 0.7
    
    def test_parse_minimal_task(self):
        """测试最小化任务输入"""
        text = "创建一个任务"
        result = self.parser.parse_task_creation(text)
        
        # 即使输入简单，也应该能提取一些信息
        assert result.raw_input == text
        assert result.confidence >= 0.0
    
    def test_parse_complex_task(self):
        """测试复杂任务"""
        text = "创建高优先级任务，标题为优化首页加载性能，描述为当前首页加载时间超过5秒，需要优化到2秒以内。开始日期为下周一，截止日期为下周五，分配给技术团队，标签：performance, frontend, optimization"
        result = self.parser.parse_task_creation(text)
        
        assert result.title is not None
        assert result.description is not None
        assert result.priority == PriorityLevel.HIGH
        assert result.start_date is not None
        assert result.due_date is not None
        assert len(result.labels) >= 2
    
    def test_confidence_calculation(self):
        """测试置信度计算"""
        # 完整信息应该获得高置信度
        text_full = "创建高优先级任务，标题为Test，截止日期明天，分配给test@example.com #bug"
        result_full = self.parser.parse_task_creation(text_full)
        
        # 不完整信息应该获得较低置信度
        text_partial = "创建一个任务"
        result_partial = self.parser.parse_task_creation(text_partial)
        
        assert result_full.confidence > result_partial.confidence
    
    def test_validate_extraction_warnings(self):
        """测试验证警告"""
        extraction = TaskExtraction(
            raw_input="test",
            title=None,
            priority=None,
            due_date=None,
            confidence=0.3
        )
        
        warnings = self.parser.validate_extraction(extraction)
        assert len(warnings) > 0
        assert any("Title" in w for w in warnings)
        assert any("Priority" in w for w in warnings)
    
    def test_parsing_notes(self):
        """测试解析说明"""
        text = "创建高优先级任务，标题为Test，分配给张三"
        result = self.parser.parse_task_creation(text)
        
        assert len(result.parsing_notes) > 0
        assert any("Priority" in note for note in result.parsing_notes)
        assert any("Assignee" in note for note in result.parsing_notes)


class TestEdgeCases:
    """边界情况测试"""
    
    def setup_method(self):
        self.parser = NLPParser()
    
    def test_empty_input(self):
        """测试空输入"""
        with pytest.raises(Exception):
            self.parser.parse_task_creation("")
    
    def test_very_long_input(self):
        """测试超长输入"""
        text = "创建任务 " * 1000
        result = self.parser.parse_task_creation(text)
        assert result is not None
    
    def test_special_characters(self):
        """测试特殊字符"""
        text = "创建任务：@#$%^&*()标题为<>{}[]测试"
        result = self.parser.parse_task_creation(text)
        assert result is not None
    
    def test_mixed_language(self):
        """测试混合语言"""
        text = "Create高优先级任务，标题为Fix bug，due明天"
        result = self.parser.parse_task_creation(text)
        assert result is not None
        assert result.priority == PriorityLevel.HIGH


# 运行测试
if __name__ == "__main__":
    pytest.main([__file__, "-v"])
