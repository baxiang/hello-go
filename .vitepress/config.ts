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
      { text: '基础语法', link: '/01-basics/01-环境与工具/01-环境搭建与工具链' },
      { text: '核心特性', link: '/02-core/01-结构体与方法' },
      { text: '并发编程', link: '/03-concurrency/01-Goroutine与GMP模型' },
      { text: '标准库', link: '/04-stdlib/01-常用标准库' },
      { text: 'Web 开发', link: '/05-web/01-Web框架详解' },
      { text: '工程实践', link: '/06-engineering/01-项目结构' },
      { text: '进阶', link: '/07-advanced/01-内存管理' },
      { text: '项目实战', link: '/08-projects/01-入门级项目' },
      { text: '云原生', link: '/09-cloud-native/01-Docker容器化进阶' },
      { text: '性能优化', link: '/10-performance/02-pprof工具详解' },
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
