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
    path: '/',
    component: ComponentCreator('/', 'e5f'),
    exact: true
  },
  {
    path: '/',
    component: ComponentCreator('/', 'aa1'),
    routes: [
      {
        path: '/',
        component: ComponentCreator('/', '519'),
        routes: [
          {
            path: '/',
            component: ComponentCreator('/', '0ad'),
            routes: [
              {
                path: '/backend/audit',
                component: ComponentCreator('/backend/audit', '01b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/auth',
                component: ComponentCreator('/backend/auth', 'b2f'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/database',
                component: ComponentCreator('/backend/database', 'e81'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/device-trust',
                component: ComponentCreator('/backend/device-trust', 'b3d'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/grpc-api-overview',
                component: ComponentCreator('/backend/grpc-api-overview', '1f7'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/health',
                component: ComponentCreator('/backend/health', '56a'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/mfa',
                component: ComponentCreator('/backend/mfa', '9cb'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/org-policy-config',
                component: ComponentCreator('/backend/org-policy-config', '25b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/organization-membership',
                component: ComponentCreator('/backend/organization-membership', '5a4'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/policy-engine',
                component: ComponentCreator('/backend/policy-engine', '73e'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/session-lifecycle',
                component: ComponentCreator('/backend/session-lifecycle', 'f85'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/sessions',
                component: ComponentCreator('/backend/sessions', 'cb4'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/telemetry',
                component: ComponentCreator('/backend/telemetry', '85f'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/backend/testing',
                component: ComponentCreator('/backend/testing', '75c'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/category/backend',
                component: ComponentCreator('/category/backend', 'cc4'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/category/frontend',
                component: ComponentCreator('/category/frontend', 'b3b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/category/operations',
                component: ComponentCreator('/category/operations', '11d'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/contributing/pending',
                component: ComponentCreator('/contributing/pending', '948'),
                exact: true
              },
              {
                path: '/frontend/architecture',
                component: ComponentCreator('/frontend/architecture', '4ab'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/frontend/dashboard',
                component: ComponentCreator('/frontend/dashboard', 'a94'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/frontend/user-browser',
                component: ComponentCreator('/frontend/user-browser', 'c22'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/intro',
                component: ComponentCreator('/intro', '4a2'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/operations/deployment',
                component: ComponentCreator('/operations/deployment', '161'),
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
    path: '*',
    component: ComponentCreator('*'),
  },
];
