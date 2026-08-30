# TalkEx Business — marketing site

Static single-page app (hash-routed) that lives at **[business.talkex.in](https://business.talkex.in)**.
The actual product dashboard is a separate Vite app at `/frontend`, deployed to
`app.business.talkex.in`.

## Structure

- **`index.html`** — one file, ~240 KB. Every route (`#/product`, `#/pricing`,
  `#/blog`, `#/docs`, `#/status`, `#/case-studies`, `#/demo`, `#/legal/*`, etc.)
  is a section swap in the same document.
- **`vercel.json`** — SPA rewrites, security headers, CSP restricted to Google
  Fonts + `app.business.talkex.in`, `/signup` and `/login` redirects to the app.

## Local preview

Any static server works — the file has no build step.

```bash
npx serve marketing
# or
python -m http.server 4000 --directory marketing
```

Then open <http://localhost:3000/> (or `:4000`).

## Deploy

```bash
cd marketing && vercel --prod
```

Point the `business.talkex.in` apex + `www.business.talkex.in` at the resulting project. The
dashboard app (Vercel project #2) sits on `app.business.talkex.in` and is deployed from
`/frontend` with its own `deploy/vercel.json`.

## Wiring to the app

- Header **Sign in** → `https://app.business.talkex.in/login`
- All **Start free** CTAs → `https://app.business.talkex.in/register`
- Footer **Book a demo** → `#/demo` (in-page booking form)
- Footer **Status** → `#/status` (mirror of api uptime)

The three redirects in `vercel.json` (`/signup`, `/login`, `/app`) preserve
short-links that predate the split.

## Editing content

Search for the section id you want to change (`id="page-pricing"`,
`id="page-blog-post-migrate"`, …). Copy is intentionally short and unpolished —
edit in place, no build.
