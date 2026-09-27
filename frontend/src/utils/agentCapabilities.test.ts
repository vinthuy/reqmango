import { describe, it, expect } from 'vitest'
import { capabilityLabels } from './agentCapabilities'

const t = (key: string) => `[${key}]`

describe('capabilityLabels', () => {
  it('folds tool names into the editor categories without repeats', () => {
    expect(
      capabilityLabels(
        ['get_issue', 'get_issue_activities', 'search_issues', 'list_states', 'list_members', 'suggest_issue_changes'],
        t,
      ),
    ).toEqual(['[agent.search]', '[agent.listResources]', '[agent.suggestChanges]'])
  })

  it('labels editor categories and legacy names', () => {
    expect(capabilityLabels(['search', 'summarize', 'triage'], t)).toEqual([
      '[agent.search]',
      '[agent.summarize]',
      '[agent.capTriage]',
    ])
  })

  it('keeps unknown capabilities readable as-is', () => {
    expect(capabilityLabels(['my_custom_tool', 'search'], t)).toEqual(['my_custom_tool', '[agent.search]'])
  })

  it('returns nothing for no capabilities', () => {
    expect(capabilityLabels(undefined, t)).toEqual([])
  })
})
