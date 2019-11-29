module.exports = {
    title: 'Goyave',
    description: 'An Elegant Golang Web Framework',
    dest: '../docs',
    base: '/goyave/',
    head: [
        ['link', { rel: 'icon', type: "image/png", sizes: "16x16", href: `/goyave_16.png` }],
        ['link', { rel: 'icon', type: "image/png", sizes: "32x32", href: `/goyave_32.png` }],
        ['link', { rel: 'icon', type: "image/png", sizes: "64x64", href: `/goyave_64.png` }],
        ['link', { rel: 'icon', type: "image/png", sizes: "128x128", href: `/goyave_128.png` }],
        ['link', { rel: 'icon', type: "image/png", sizes: "256x256", href: `/goyave_256.png` }],
        ['link', { rel: 'icon', type: "image/png", sizes: "512x512", href: `/goyave_512.png` }],
    ],
    themeConfig: {
        repo: 'System-Glitch/goyave',
        editLinks: true,
        docsDir: 'docs_src/src',
        smoothScroll: true,
        activeHeaderLinks: false,
        logo: '/goyave_64.png',
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
        ['vuepress-plugin-container', {
            type: 'img-row',
            before: (img) => `<div class="img-row left">${img}<div class="row-content">`,
            after: '</div></div>',
        }],
        ['vuepress-plugin-container', {
            type: 'img-row-right',
            before: (img) => `<div class="img-row right">${img}<div class="row-content">`,
            after: '</div></div>',
        }],
        ['vuepress-plugin-container', {
            type: 'vue',
            before: '<pre class="vue-container"><code>',
            after: '</code></pre>',
        }],
        ['vuepress-plugin-container', {
            type: 'table',
            before: '<div class="table">',
            after: '</div>',
        }]
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
                'changelog',
                'installation',
                'upgrade-guide',
                'configuration',
                'architecture-concepts',
                'deployment',
                'contribution-guide',
            ]
        },
        {
            title: 'The Basics',
            collapsable: true,
            children: [
                'basics/routing',
                'basics/middleware',
                'basics/requests',
                'basics/controllers',
                'basics/responses',
                'basics/database',
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
                'advanced/multi-services',
            ]
        }
    ]
}