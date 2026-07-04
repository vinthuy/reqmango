/** Project Page Tab types */

export interface ProjectPageTab {
  id: number
  project_id: number
  owner_id: number
  name: string
  icon: string
  tab_type: 'issues' | 'cycles' | 'modules' | 'updates' | 'pages' | 'settings' | 'analytics' | 'dashboards' | 'roadmap' | 'releases' | 'custom'
  route_key: string
  target_type: 'saved_view' | 'url' | ''
  target_id?: number
  target_url: string
  visible: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

/** Built-in tab definitions for defaults and the config panel */
export interface BuiltInTabDef {
  tab_type: ProjectPageTab['tab_type']
  name: string
  icon: string
  route_key: string
}
