import { defineConfig } from 'vitepress'
import sidebar from './sidebar'

const isProduction = process.env.NODE_ENV === 'production' || process.argv.includes('build')

export default defineConfig({
  title: '从零开始系统学习 Go 编程',
  description: 'Go 全栈学习路线教程',
  lang: 'zh-CN',
  base: isProduction ? '/hello-go/' : '/',
  cleanUrls: true,
  lastUpdated: true,
  ignoreDeadLinks: true,

  themeConfig: {
    nav: [
      { text: '首页', link: '/' },
      { text: '语言基础', link: '/01-语言基础/01-环境与工具/01-环境搭建与工具链' },
      { text: '进阶编程', link: '/02-进阶编程/01-Goroutine与GMP模型' },
      { text: 'Web与工程', link: '/03-Web与工程/01-Web框架详解' },
      { text: '深度进阶', link: '/04-深度进阶/01-内存管理' },
      { text: '项目实战', link: '/05-实战项目/kratos/' },
      { text: '云原生', link: '/06-云原生/01-Docker容器化进阶' },
    ],

    sidebar,

    search: {
      provider: 'local',
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/baxiang/hello-go' },
    ],

    outline: {
      label: '本页目录',
      level: [2, 3],
    },

    docFooter: {
      prev: '上一篇',
      next: '下一篇',
    },

    lastUpdated: {
      text: '最后更新',
      formatOptions: {
        dateStyle: 'medium',
        timeStyle: 'short',
      },
    },
  },

  markdown: {
    lineNumbers: true,
  },

  srcExclude: [
    '**/node_modules/**',
    '**/.vitepress/**',
    'docs/**',
    'AGENTS.md',
    'CLAUDE.md',
    'QWEN.md',
    'RENAME_PLAN.md',
    'TUTORIAL_REVIEW_REPORT.md',
    '.github/**',
  ],
})
