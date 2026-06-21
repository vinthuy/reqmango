/**
 * Custom Field Types - 自定义字段类型定义
 */

// ==================== Enums ====================

export enum CustomFieldTypeEnum {
  TEXT = 'text',
  NUMBER = 'number',
  DROPDOWN = 'dropdown',
  BOOLEAN = 'boolean',
  DATE = 'date',
  MEMBER = 'member',
  URL = 'url'
}

export enum TextTypeEnum {
  SINGLE = 'single',
  PARAGRAPH = 'paragraph',
  READONLY = 'readonly'
}

// ==================== Custom Field Option ====================

export interface CustomFieldOption {
  id: number
  value: string
  color?: string
  sequence: number
  is_default: boolean
  is_active: boolean
}

export interface CustomFieldOptionCreate {
  value: string
  color?: string
  sequence?: number
  is_default?: boolean
  is_active?: boolean
}

export interface CustomFieldOptionUpdate {
  value?: string
  color?: string
  sequence?: number
  is_default?: boolean
  is_active?: boolean
}

// ==================== Custom Field ====================

export interface CustomField {
  id: number
  name: string
  description?: string
  field_type: CustomFieldTypeEnum
  is_required: boolean
  is_readonly: boolean
  is_active: boolean
  is_unique?: boolean
  
  // 文本类型属性
  text_type?: TextTypeEnum
  placeholder?: string
  
  // 数字类型属性
  number_default?: number
  number_min?: number
  number_max?: number
  
  // 下拉类型属性
  is_multi_select: boolean
  
  // 日期类型属性
  date_format?: string
  
  // 序列号
  sequence: number
  
  // 默认值
  default_value?: any
  
  // 关联关系
  workspace_id: number
  project_id?: number
  issue_type_id?: number
  
  // 选项列表
  options: CustomFieldOption[]
  
  // 时间戳
  created_at: string
  updated_at: string
}

export interface CustomFieldCreate {
  name: string
  description?: string
  field_type: CustomFieldTypeEnum
  is_required?: boolean
  is_readonly?: boolean
  is_active?: boolean
  is_unique?: boolean
  default_value?: any
  min_value?: number
  max_value?: number
  min_length?: number
  max_length?: number
  
  text_type?: TextTypeEnum
  placeholder?: string
  
  number_default?: number
  number_min?: number
  number_max?: number
  
  is_multi_select?: boolean
  
  date_format?: string
  
  sequence?: number
  
  project_id?: number
  issue_type_id?: number
  
  options?: CustomFieldOptionCreate[]
}

export interface CustomFieldUpdate {
  name?: string
  description?: string
  is_required?: boolean
  is_readonly?: boolean
  is_active?: boolean
  is_unique?: boolean
  default_value?: any
  min_value?: number
  max_value?: number
  min_length?: number
  max_length?: number
  
  text_type?: TextTypeEnum
  placeholder?: string
  
  number_default?: number
  number_min?: number
  number_max?: number
  
  is_multi_select?: boolean
  
  date_format?: string
  
  sequence?: number
}

export interface CustomFieldLite {
  id: number
  name: string
  field_type: string
  is_required: boolean
  is_readonly: boolean
  options: CustomFieldOption[]
  text_type?: string
  placeholder?: string
  number_min?: number
  number_max?: number
  is_multi_select?: boolean
}

// ==================== Issue Custom Field Value ====================

export interface IssueCustomFieldValue {
  id: number
  issue_id: number
  field_id: number
  field_name: string
  field_type: string
  
  // 显示值（处理后）
  value?: any
  
  // 原始值
  text_value?: string
  number_value?: number
  boolean_value?: boolean
  date_value?: string
  url_value?: string
  json_value?: number[]
}

export interface IssueCustomFieldValueCreate {
  issue_id: number
  field_id: number
  text_value?: string
  number_value?: number
  boolean_value?: boolean
  date_value?: string
  url_value?: string
  json_value?: number[]
}

export interface IssueCustomFieldValueUpdate {
  text_value?: string
  number_value?: number
  boolean_value?: boolean
  date_value?: string
  url_value?: string
  json_value?: number[]
}

export interface BulkCustomFieldValueUpdate {
  issue_id: number
  values: IssueCustomFieldValueUpdate[]
}

// ==================== Custom Field with Values ====================

export interface CustomFieldWithValues {
  field: CustomFieldLite
  value?: any
}

export interface IssueCustomFieldsResponse {
  issue_id: number
  fields: CustomFieldWithValues[]
}

// ==================== Validation ====================

export interface CustomFieldValidationRequest {
  field_id: number
  value: any
}

export interface CustomFieldValidationResponse {
  is_valid: boolean
  errors: string[]
  field_id: number
}

// ==================== Helper Functions ====================

/**
 * 获取字段类型的显示名称
 */
export function getFieldTypeName(type: CustomFieldTypeEnum): string {
  const names: Record<CustomFieldTypeEnum, string> = {
    [CustomFieldTypeEnum.TEXT]: '文本',
    [CustomFieldTypeEnum.NUMBER]: '数字',
    [CustomFieldTypeEnum.DROPDOWN]: '下拉选择',
    [CustomFieldTypeEnum.BOOLEAN]: '布尔值',
    [CustomFieldTypeEnum.DATE]: '日期',
    [CustomFieldTypeEnum.MEMBER]: '成员选择',
    [CustomFieldTypeEnum.URL]: '链接'
  }
  return names[type] || type
}

/**
 * 根据字段类型获取字段值的输入类型
 */
export function getInputType(fieldType: CustomFieldTypeEnum): string {
  const types: Record<CustomFieldTypeEnum, string> = {
    [CustomFieldTypeEnum.TEXT]: 'text',
    [CustomFieldTypeEnum.NUMBER]: 'number',
    [CustomFieldTypeEnum.DROPDOWN]: 'select',
    [CustomFieldTypeEnum.BOOLEAN]: 'checkbox',
    [CustomFieldTypeEnum.DATE]: 'date',
    [CustomFieldTypeEnum.MEMBER]: 'select',
    [CustomFieldTypeEnum.URL]: 'url'
  }
  return types[fieldType] || 'text'
}

/**
 * 创建空字段值对象
 */
export function createEmptyFieldValue(field: CustomField): IssueCustomFieldValueUpdate {
  const value: IssueCustomFieldValueUpdate = {}
  
  switch (field.field_type) {
    case CustomFieldTypeEnum.TEXT:
      value.text_value = ''
      break
    case CustomFieldTypeEnum.NUMBER:
      value.number_value = field.number_default ?? 0
      break
    case CustomFieldTypeEnum.BOOLEAN:
      value.boolean_value = false
      break
    case CustomFieldTypeEnum.DATE:
      value.date_value = ''
      break
    case CustomFieldTypeEnum.URL:
      value.url_value = ''
      break
    case CustomFieldTypeEnum.DROPDOWN:
      value.json_value = []
      break
    case CustomFieldTypeEnum.MEMBER:
      value.json_value = []
      break
  }
  
  return value
}

/**
 * 格式化字段值用于显示
 */
export function formatFieldValue(field: CustomField, value: IssueCustomFieldValueUpdate): string {
  switch (field.field_type) {
    case CustomFieldTypeEnum.TEXT:
      return value.text_value || ''
    case CustomFieldTypeEnum.NUMBER:
      return value.number_value?.toString() || ''
    case CustomFieldTypeEnum.BOOLEAN:
      return value.boolean_value ? '是' : '否'
    case CustomFieldTypeEnum.DATE:
      return value.date_value || ''
    case CustomFieldTypeEnum.URL:
      return value.url_value || ''
    case CustomFieldTypeEnum.DROPDOWN:
      if (value.json_value && field.options) {
        const optionMap = new Map(field.options.map(o => [o.id, o.value]))
        return value.json_value.map(id => optionMap.get(id) || id).join(', ')
      }
      return ''
    case CustomFieldTypeEnum.MEMBER:
      return value.json_value?.join(', ') || ''
    default:
      return ''
  }
}