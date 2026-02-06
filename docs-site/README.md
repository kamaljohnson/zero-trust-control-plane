# Zero Trust Control Plane — Docs site

Documentation is built with [Docusaurus](https://docusaurus.io). The frontend app links to this site via `NEXT_PUBLIC_DOCS_URL` (e.g. `http://localhost:3001` in development).

## Requirements

- Node.js >= 20.0

## Commands

```bash
npm install
npm run start    # dev server, default port 3001
npm run build    # production build
npm run serve    # serve production build
```

## Structure

- **docs/** — Markdown source (intro, backend/*, contributing/*).
- **sidebars.ts** — Sidebar order and categories (Getting started, Backend, Contributing).
- **static/** — Static assets.
