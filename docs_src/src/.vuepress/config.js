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
                    '/guide/': getGuideSidebar('Guide'),
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

function getGuideSidebar (groupA) {
    return [
        {
            title: groupA,
            collapsable: false,
            children: [
                '',
                'getting-started',
                'the-basics',
                'authentication',
                'database',
                'testing',
                'deployment',
                // TODO add other pages
            ]
        }
    ]
}