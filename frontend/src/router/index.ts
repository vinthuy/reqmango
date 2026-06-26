import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('@/views/Home.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue')
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/Register.vue')
    },
    {
      path: '/workspace/:slug',
      name: 'Workspace',
      component: () => import('@/views/Workspace.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id',
      name: 'Project',
      component: () => import('@/views/Project.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/settings',
      name: 'WorkspaceSettings',
      component: () => import('@/views/WorkspaceSettings.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/settings',
      name: 'ProjectSettings',
      component: () => import('@/views/ProjectSettings.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/settings/workflows/:workflowId',
      name: 'WorkflowDetail',
      component: () => import('@/views/WorkflowDetail.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/issues/:issueId',
      name: 'IssueDetail',
      component: () => import('@/views/IssueDetail.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/issues/:issueId',
      name: 'IssueDetailOld',
      component: () => import('@/views/IssueDetail.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/issues/new',
      name: 'IssueCreate',
      component: () => import('@/views/IssueCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/cycles/new',
      name: 'CycleCreate',
      component: () => import('@/views/CycleCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/cycles/:cycleId',
      name: 'CycleDetail',
      component: () => import('@/views/CycleDetail.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/custom-fields',
      name: 'CustomFields',
      component: () => import('@/views/CustomFields.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/issue-types',
      name: 'IssueTypeList',
      component: () => import('@/views/IssueTypeList.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/intake/:projectId',
      name: 'IntakeForm',
      component: () => import('@/views/IntakeForm.vue'),
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/issues/new',
      name: 'IssueCreateOld',
      component: () => import('@/views/IssueCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/cycles/new',
      name: 'CycleCreateOld',
      component: () => import('@/views/CycleCreate.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspaces/:workspaceId/projects/:projectId/cycles/:cycleId',
      name: 'CycleDetailOld',
      component: () => import('@/views/CycleDetail.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/workspace/:slug/project/:id/pages',
      name: 'ProjectPages',
      component: () => import('@/views/ProjectPages.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth) {
    if (!authStore.token) {
      return { name: 'Login' }
    }
    
    if (!authStore.user) {
      try {
        await authStore.fetchUser()
      } catch {
        authStore.logout()
        return { name: 'Login' }
      }
    }
  }
  
  if ((to.name === 'Login' || to.name === 'Register') && authStore.token) {
    if (!authStore.user) {
      try {
        await authStore.fetchUser()
      } catch {
        authStore.logout()
        return
      }
    }
    return { name: 'Home' }
  }
})

export default router