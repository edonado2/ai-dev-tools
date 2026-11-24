// docs/.vitepress/config.mjs
import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'ai-dev-tools',
  description: 'AI-Powered Developer Tools',
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      {
        text: 'Tools',
        items: [
          { text: 'Commit', link: '/tools/commit' },
          { text: 'Review', link: '/tools/review' },
          { text: 'TestGen', link: '/tools/testgen' },
          { text: 'Changelog', link: '/tools/changelog' },
        ],
      },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'GitHub', link: 'https://github.com/ernedo/ai-dev-tools' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Providers', link: '/guide/providers' },
            { text: 'GitHub Integration', link: '/guide/github' },
          ],
        },
      ],
      '/tools/': [
        {
          text: 'Tools',
          items: [
            { text: 'Commit Generator', link: '/tools/commit' },
            { text: 'Code Reviewer', link: '/tools/review' },
            { text: 'Test Generator', link: '/tools/testgen' },
            { text: 'Changelog Generator', link: '/tools/changelog' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'CLI Reference', link: '/reference/cli' },
            { text: 'Config Schema', link: '/reference/config-schema' },
            { text: 'Prompt Templates', link: '/reference/templates' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/ernedo/ai-dev-tools' },
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present ernedo',
    },
  },
})