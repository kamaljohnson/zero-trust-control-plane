import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/zero-trust-control-plane/__docusaurus/debug',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug', 'da8'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/config',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/config', 'aed'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/content',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/content', 'cac'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/globalData',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/globalData', '98a'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/metadata',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/metadata', '7a4'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/registry',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/registry', '322'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/__docusaurus/debug/routes',
    component: ComponentCreator('/zero-trust-control-plane/__docusaurus/debug/routes', '3e5'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/',
    component: ComponentCreator('/zero-trust-control-plane/', '35f'),
    exact: true
  },
  {
    path: '/zero-trust-control-plane/',
    component: ComponentCreator('/zero-trust-control-plane/', 'cf1'),
    routes: [
      {
        path: '/zero-trust-control-plane/',
        component: ComponentCreator('/zero-trust-control-plane/', '270'),
        routes: [
          {
            path: '/zero-trust-control-plane/',
            component: ComponentCreator('/zero-trust-control-plane/', 'e5b'),
            routes: [
              {
                path: '/zero-trust-control-plane/backend/audit',
                component: ComponentCreator('/zero-trust-control-plane/backend/audit', '3e0'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/auth',
                component: ComponentCreator('/zero-trust-control-plane/backend/auth', '547'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/database',
                component: ComponentCreator('/zero-trust-control-plane/backend/database', '280'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/device-trust',
                component: ComponentCreator('/zero-trust-control-plane/backend/device-trust', '06a'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/grpc-api-overview',
                component: ComponentCreator('/zero-trust-control-plane/backend/grpc-api-overview', '84b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/health',
                component: ComponentCreator('/zero-trust-control-plane/backend/health', 'f0c'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/mfa',
                component: ComponentCreator('/zero-trust-control-plane/backend/mfa', '6fa'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/org-policy-config',
                component: ComponentCreator('/zero-trust-control-plane/backend/org-policy-config', '313'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/organization-membership',
                component: ComponentCreator('/zero-trust-control-plane/backend/organization-membership', 'd83'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/policy-engine',
                component: ComponentCreator('/zero-trust-control-plane/backend/policy-engine', '250'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/session-lifecycle',
                component: ComponentCreator('/zero-trust-control-plane/backend/session-lifecycle', 'd8b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/sessions',
                component: ComponentCreator('/zero-trust-control-plane/backend/sessions', '4b1'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/telemetry',
                component: ComponentCreator('/zero-trust-control-plane/backend/telemetry', '8d8'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/backend/testing',
                component: ComponentCreator('/zero-trust-control-plane/backend/testing', '8a1'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/category/backend',
                component: ComponentCreator('/zero-trust-control-plane/category/backend', '648'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/category/frontend',
                component: ComponentCreator('/zero-trust-control-plane/category/frontend', '6e0'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/category/operations',
                component: ComponentCreator('/zero-trust-control-plane/category/operations', '71b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/contributing/pending',
                component: ComponentCreator('/zero-trust-control-plane/contributing/pending', '4f1'),
                exact: true
              },
              {
                path: '/zero-trust-control-plane/frontend/architecture',
                component: ComponentCreator('/zero-trust-control-plane/frontend/architecture', '107'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/frontend/dashboard',
                component: ComponentCreator('/zero-trust-control-plane/frontend/dashboard', 'c50'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/frontend/user-browser',
                component: ComponentCreator('/zero-trust-control-plane/frontend/user-browser', '330'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/intro',
                component: ComponentCreator('/zero-trust-control-plane/intro', '19f'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/zero-trust-control-plane/operations/deployment',
                component: ComponentCreator('/zero-trust-control-plane/operations/deployment', '75f'),
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
