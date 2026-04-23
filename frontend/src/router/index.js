import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue')
  },
  {
    path: '/style-guide',
    name: 'StyleGuide',
    component: () => import('../views/StyleGuide.vue')
  },
  {
    path: '/prompts',
    name: 'PromptList',
    component: () => import('../views/PromptList.vue')
  },
  {
    path: '/prompts/:id',
    name: 'PromptEditor',
    component: () => import('../views/PromptEditor.vue')
  },
  {
    path: '/prompts/:id/versions',
    name: 'VersionHistory',
    component: () => import('../views/VersionHistory.vue')
  },
  {
    path: '/prompts/:id/compare',
    name: 'VersionCompare',
    component: () => import('../views/VersionCompare.vue')
  },
  {
    path: '/prompts/:id/translate',
    name: 'PromptTranslate',
    component: () => import('../views/TranslationCompare.vue')
  },
  {
    path: '/prompts/:id/test',
    name: 'PromptTester',
    component: () => import('../views/PromptTester.vue')
  },
  {
    path: '/prompts/:id/test-compare',
    name: 'TestCompare',
    component: () => import('../views/TestCompare.vue')
  },
  {
    path: '/prompts/:id/optimize',
    name: 'OptimizePrompt',
    component: () => import('../views/OptimizePrompt.vue')
  },
  {
    path: '/prompts/:id/analytics',
    name: 'TestAnalytics',
    component: () => import('../views/TestAnalytics.vue')
  },

  // Skills routes
  {
    path: '/skills',
    name: 'SkillList',
    component: () => import('../views/SkillList.vue')
  },
  {
    path: '/skills/:id',
    name: 'SkillEditor',
    component: () => import('../views/SkillEditor.vue')
  },
  {
    path: '/skills/:id/translate',
    name: 'SkillTranslate',
    component: () => import('../views/TranslationCompare.vue')
  },

  // Agents routes
  {
    path: '/agents',
    name: 'AgentList',
    component: () => import('../views/AgentList.vue')
  },
  {
    path: '/agents/:id',
    name: 'AgentEditor',
    component: () => import('../views/AgentEditor.vue')
  },
  {
    path: '/agents/:id/translate',
    name: 'AgentTranslate',
    component: () => import('../views/TranslationCompare.vue')
  },

  // Settings & Activity
  {
    path: '/activity',
    name: 'ActivityLog',
    component: () => import('../views/ActivityLog.vue')
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue')
  },

  // API Documentation
  {
    path: '/api-docs',
    name: 'ApiDocs',
    component: () => import('../views/ApiDocs.vue')
  },

  // A/B Testing
  {
    path: '/ab-tests',
    name: 'ABTestList',
    component: () => import('../views/ABTestList.vue')
  },
  {
    path: '/ab-tests/:id',
    name: 'ABTestDetail',
    component: () => import('../views/ABTestDetail.vue')
  },

  // Team Collaboration
  {
    path: '/teams',
    name: 'TeamList',
    component: () => import('../views/TeamList.vue')
  },
  {
    path: '/teams/:id/members',
    name: 'TeamMemberList',
    component: () => import('../views/TeamMemberList.vue')
  },
  {
    path: '/teams/:id/settings',
    name: 'TeamSettings',
    component: () => import('../views/TeamSettings.vue')
  },

  // Template Marketplace
  {
    path: '/templates',
    name: 'TemplateMarketplace',
    component: () => import('../views/TemplateMarketplace.vue')
  },
  {
    path: '/templates/:id',
    name: 'TemplateDetail',
    component: () => import('../views/TemplateDetail.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
