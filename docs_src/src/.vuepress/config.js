module.exports = {
    title: 'Goyave',
    description: 'An all-in-one elegant Go web framework',
    dest: '../docs',
    base: '/goyave/',
    head: [
        ['link', { rel: 'icon', href: `/logo.svg` }],
    ],
    themeConfig: {
        repo: 'System-Glitch/goyave',
        editLinks: true,
        docsDir: 'docs_src/src',
        smoothScroll: true,
        locales: {
            '/': {
                label: 'English',
                selectText: 'Languages',
                editLinkText: 'Edit this page on GitHub',
                lastUpdated: 'Last Updated',
                nav: require('./nav/en'),
                sidebar: {
                    '/guide/': getGuideSidebar(),
                }
            }
        }
    },
    plugins: [
        ['@vuepress/back-to-top', true],
    ],
    extraWatchFiles: [
        '.vuepress/nav/en.js',
    ]
    
}

function getGuideSidebar () {
    return [
        {
            title: 'Guide',
            collapsable: true,
            children: [
                '',
                'installation',
                'upgrade-guide',
                'configuration',
                'contribution-guide',
                'architecture-concepts',
                'deployment',
            ]
        },
        {
            title: 'The Basics',
            collapsable: true,
            children: [
                'basics/routing',
                'basics/middlewares',
                'basics/requests',
                'basics/controllers',
                'basics/database',
                'basics/responses',
                'basics/validation',
            ]
        },
        {
            title: 'Advanced',
            collapsable: true,
            children: [
                'advanced/helpers',
                'advanced/authentication',
                'advanced/localization',
                'advanced/testing',
                'advanced/plugins',
            ]
        }
    ]
}