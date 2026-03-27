import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue')
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
    path: '/prompts/:id/optimize',
    name: 'OptimizePrompt',
    component: () => import('../views/OptimizePrompt.vue')
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
