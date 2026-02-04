import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/__docusaurus/debug',
    component: ComponentCreator('/__docusaurus/debug', '5ff'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/config',
    component: ComponentCreator('/__docusaurus/debug/config', '5ba'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/content',
    component: ComponentCreator('/__docusaurus/debug/content', 'a2b'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/globalData',
    component: ComponentCreator('/__docusaurus/debug/globalData', 'c3c'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/metadata',
    component: ComponentCreator('/__docusaurus/debug/metadata', '156'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/registry',
    component: ComponentCreator('/__docusaurus/debug/registry', '88c'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/routes',
    component: ComponentCreator('/__docusaurus/debug/routes', '000'),
    exact: true
  },
  {
    path: '/docs',
    component: ComponentCreator('/docs', 'ae0'),
    routes: [
      {
        path: '/docs',
        component: ComponentCreator('/docs', '2de'),
        routes: [
          {
            path: '/docs',
            component: ComponentCreator('/docs', 'c1d'),
            routes: [
              {
                path: '/docs/backend/audit',
                component: ComponentCreator('/docs/backend/audit', '642'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/auth',
                component: ComponentCreator('/docs/backend/auth', 'af4'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/database',
                component: ComponentCreator('/docs/backend/database', '753'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/device-trust',
                component: ComponentCreator('/docs/backend/device-trust', 'ac5'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/grpc-api-overview',
                component: ComponentCreator('/docs/backend/grpc-api-overview', 'c64'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/health',
                component: ComponentCreator('/docs/backend/health', 'aca'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/mfa',
                component: ComponentCreator('/docs/backend/mfa', '775'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/org-policy-config',
                component: ComponentCreator('/docs/backend/org-policy-config', '42a'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/organization-membership',
                component: ComponentCreator('/docs/backend/organization-membership', '53e'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/policy-engine',
                component: ComponentCreator('/docs/backend/policy-engine', '0d3'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/session-lifecycle',
                component: ComponentCreator('/docs/backend/session-lifecycle', 'c43'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/sessions',
                component: ComponentCreator('/docs/backend/sessions', 'd82'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/telemetry',
                component: ComponentCreator('/docs/backend/telemetry', 'cfd'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/backend/testing',
                component: ComponentCreator('/docs/backend/testing', 'c7e'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/category/backend',
                component: ComponentCreator('/docs/category/backend', '145'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/category/frontend',
                component: ComponentCreator('/docs/category/frontend', '7e5'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/category/operations',
                component: ComponentCreator('/docs/category/operations', '15e'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/contributing/pending',
                component: ComponentCreator('/docs/contributing/pending', '9b4'),
                exact: true
              },
              {
                path: '/docs/frontend/architecture',
                component: ComponentCreator('/docs/frontend/architecture', '0ae'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/frontend/dashboard',
                component: ComponentCreator('/docs/frontend/dashboard', 'd4a'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/frontend/user-browser',
                component: ComponentCreator('/docs/frontend/user-browser', '894'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/intro',
                component: ComponentCreator('/docs/intro', '058'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/docs/operations/deployment',
                component: ComponentCreator('/docs/operations/deployment', 'b82'),
                exact: true,
                sidebar: "docsSidebar"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    path: '/',
    component: ComponentCreator('/', 'e5f'),
    exact: true
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
