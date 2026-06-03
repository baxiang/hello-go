import { generateSidebar } from 'vitepress-sidebar'

export default generateSidebar([
  {
    documentRootPath: '/01-basics',
    scanStartPath: '',
    resolvePath: '/01-basics/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/02-advanced',
    scanStartPath: '',
    resolvePath: '/02-advanced/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/03-web',
    scanStartPath: '',
    resolvePath: '/03-web/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/04-deep-dive',
    scanStartPath: '',
    resolvePath: '/04-deep-dive/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/05-projects',
    scanStartPath: '',
    resolvePath: '/05-projects/',
    useTitleFromFileHeading: true,
    collapsed: true,
    depthLimit: 1,
    sortMenusOrderByDescending: false,
  },
  {
    documentRootPath: '/06-cloud-native',
    scanStartPath: '',
    resolvePath: '/06-cloud-native/',
    useTitleFromFileHeading: true,
    collapsed: true,
    sortMenusOrderByDescending: false,
  },
])
