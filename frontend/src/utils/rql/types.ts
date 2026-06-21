// frontend/src/utils/rql/types.ts

// Token 类型
export enum TokenType {
  IDENTIFIER = 'IDENTIFIER',   // 字段名
  STRING = 'STRING',           // 字符串
  NUMBER = 'NUMBER',           // 数字
  DATE = 'DATE',              // 日期 2026-01-01
  DATETIME = 'DATETIME',       // 日期时间
  OPERATOR = 'OPERATOR',      // =, !=, >, <, >=, <=
  LIKE = 'LIKE',              // LIKE
  IN = 'IN',                  // IN
  AND = 'AND',                // AND
  OR = 'OR',                  // OR
  NOT = 'NOT',                // NOT
  LPAREN = 'LPAREN',          // (
  RPAREN = 'RPAREN',          // )
  COMMA = 'COMMA',            // ,
  EOF = 'EOF'
}

export interface Token {
  type: TokenType
  value: string
  position: number
}

// AST 节点类型
export type ASTNodeType = 'BinaryExpr' | 'Comparison' | 'InExpr' | 'LikeExpr'

export interface ASTNode {
  type: ASTNodeType
}

export interface BinaryExpr extends ASTNode {
  type: 'BinaryExpr'
  operator: 'AND' | 'OR'
  left: ASTNode
  right: ASTNode
}

export interface Comparison extends ASTNode {
  type: 'Comparison'
  field: string
  operator: '=' | '!=' | '>' | '<' | '>=' | '<='
  value: string | number | Date
}

export interface LikeExpr extends ASTNode {
  type: 'LikeExpr'
  field: string
  value: string
}

export interface InExpr extends ASTNode {
  type: 'InExpr'
  field: string
  values: (string | number)[]
}

// 查询参数
export interface IssueQueryParams {
  search?: string
  state_id?: number
  priority?: string
  assignee_id?: number
  reporter_id?: number
  cycle_id?: number
  module_id?: number
  label_ids?: number[]
  start_date?: string
  end_date?: string
  due_date_start?: string
  due_date_end?: string
  created_after?: string
  created_before?: string
  updated_after?: string
  updated_before?: string
}

// 字段定义
export interface FieldDefinition {
  name: string
  type: 'string' | 'number' | 'date' | 'datetime' | 'enum' | 'user'
  enumValues?: string[]
  label: string
}

// 预定义字段
export const ISSUE_FIELDS: FieldDefinition[] = [
  { name: 'id', type: 'number', label: 'ID' },
  { name: 'sequence_id', type: 'string', label: '编号' },
  { name: 'name', type: 'string', label: '标题' },
  { name: 'description', type: 'string', label: '描述' },
  { name: 'state', type: 'string', label: '状态' },
  { name: 'priority', type: 'enum', enumValues: ['urgent', 'high', 'medium', 'low', 'none'], label: '优先级' },
  { name: 'assignee', type: 'user', label: '负责人' },
  { name: 'reporter', type: 'user', label: '创建者' },
  { name: 'label', type: 'string', label: '标签' },
  { name: 'cycle', type: 'string', label: '周期' },
  { name: 'module', type: 'string', label: '模块' },
  { name: 'created_at', type: 'datetime', label: '创建时间' },
  { name: 'updated_at', type: 'datetime', label: '更新时间' },
  { name: 'due_date', type: 'date', label: '截止日期' }
]

// 解析错误
export interface ParseError {
  position: number
  message: string
}

// Lexer 错误
export interface LexerError {
  position: number
  message: string
}
