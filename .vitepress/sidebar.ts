import { generateSidebar } from 'vitepress-sidebar'

export default generateSidebar([
  {
    documentRootPath: '/01-语言基础',
    scanStartPath: '',
    resolvePath: '/01-语言基础/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/02-进阶编程',
    scanStartPath: '',
    resolvePath: '/02-进阶编程/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/03-Web与工程',
    scanStartPath: '',
    resolvePath: '/03-Web与工程/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/04-深度进阶',
    scanStartPath: '',
    resolvePath: '/04-深度进阶/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/05-实战项目',
    scanStartPath: '',
    resolvePath: '/05-实战项目/',
    useTitleFromFileHeading: true,
    collapsed: true,
    depthLimit: 1,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/06-云原生',
    scanStartPath: '',
    resolvePath: '/06-云原生/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
])
