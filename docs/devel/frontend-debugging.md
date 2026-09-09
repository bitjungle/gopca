# Debugging the Desktop Frontends

## Purpose

This document exists because of #436, where a table-of-contents click did nothing and it took three attempts to find out why. Every attempt was wrong in the same way, and the mistakes are cheap to avoid once named.

The bug itself was ordinary: heading `id` attributes did not match the ids the contents list asked for. What made it expensive was the investigation.

---

## 1. Make the failure observable before you change any behaviour

The click handler contained this:

```ts
if (!container || !target || !container.contains(target)) return;
```

Three distinct failures, all silent. A dead click was indistinguishable from a dead handler, an unmounted container, a missing heading, or a scroll that ran and did nothing — so there was nothing to reason from, and reasoning is what filled the gap.

Adding one line turned two rounds of guessing into an immediate answer:

```
[DocumentationViewer] cannot scroll to "1-introduction-…": no heading has that id
```

**Rule:** when something in the interface does nothing, the first change should make it say why, not attempt a fix. A guard that can fail three ways should name which way it failed.

---

## 2. Reproduce the app's *mount*, not just its markup

The failure was reproduced in a real browser, driven with Puppeteer, using the app's own stylesheets and a faithful copy of `App.tsx`'s layout — and it **worked every time**. The jsdom component tests passed too.

Both differed from the running app in one respect that was never checked: `main.tsx` mounts inside `<React.StrictMode>`.

```tsx
root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
);
```

StrictMode invokes render functions twice in development, deliberately, to surface side effects during render. The renderer numbered duplicate headings with a `Map` it mutated while building the tree, so the second pass saw every heading as a repeat of itself and appended `-2` to each id. The contents list, built in a single pass over the markdown source, kept asking for the first.

**Rule:** component tests in `packages/ui-components` render through `StrictMode`, because both applications do. A test that renders a component directly is testing a configuration nothing ships.

**Rule:** when a browser reproduction disagrees with the app, the difference is in what you did not copy. Check the entry point before the engine, the CSS, or the build.

---

## 3. Never mutate state while rendering

The specific anti-pattern:

```tsx
// Wrong: a side effect during render.
const counts = new Map<string, number>();
h2: ({ children }) => {
  const n = counts.get(base) ?? 0;
  counts.set(base, n + 1);          // <- mutation during render
  return <h2 id={n === 0 ? base : `${base}-${n + 1}`}>{children}</h2>;
}
```

React makes no promise about how many times it calls a render function. Anything derived by counting as the tree is built will differ between renders. Derive it from the input instead — here, from each heading's line in the markdown source.

---

## 4. One value, one derivation

The heading ids were computed twice: once by scanning raw markdown lines for the contents list, and once from rendered children while building the tree. Both used the same slug function, and they still diverged — the second was stateful and the first was not.

**Rule:** when two parts of the interface must agree on a value, they call the same function. Two implementations that "obviously" agree are two implementations that will eventually disagree, and the disagreement is invisible until something looks up the result of one using the other.

---

## 5. The shared package is compiled

`packages/ui-components` is built to `dist/`, which is gitignored, and both applications import it from there. Editing its source changes nothing in a running app until it is rebuilt.

`pca-dev`, `pca-build`, `csv-dev` and `csv-build` now depend on `build-ui`, so this cannot happen silently. If you build an app some other way, rebuild the package first:

```bash
npm run build-ui
```

A running `wails dev` session does **not** pick up a rebuild of the shared package: its `dist/` sits outside the frontend directory that Vite watches. Restart the dev server after rebuilding.

---

## Reproducing a frontend problem outside the app

The desktop apps need the Wails runtime, so they cannot be opened in an ordinary browser — `#root` stays empty. Components that only need statically served files can be exercised through the Vite dev server:

```bash
cd cmd/gopca-desktop/frontend
npx vite --port 5199
```

Then add a temporary entry point that mounts the component **inside `StrictMode`**, with the app's stylesheets:

```tsx
// toc-probe.tsx, served at /toc-probe.html — delete when finished
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './src/style.css';
import '@gopca/ui-components/dist/ui-components.css';
import { DocumentationViewer, ThemeProvider } from '@gopca/ui-components';

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider>
            <DocumentationViewer isOpen onClose={() => {}}
                title="Docs" markdownPath="/docs/intro_to_pca.md" />
        </ThemeProvider>
    </StrictMode>
);
```

This is a debugging aid, not a test. Anything worth keeping belongs in a test file; delete the probe before committing.

---

## The app's own console

`wails dev` builds with devtools enabled: right-click in the window and choose *Inspect Element*. Two lines answer most layout and navigation questions:

```js
const c = document.querySelector('[data-testid=documentation-scroll]');
console.log(c?.scrollHeight, c?.clientHeight, document.querySelectorAll('h2[id]').length);
```

If `scrollHeight === clientHeight`, an element with `overflow-y-auto` is not scrolling — usually a flex child without `min-h-0`, which defaults to `min-height: auto` and refuses to shrink below its content. Chromium often resolves this once `overflow` is set; the macOS WebKit view does not.

Ignore the `WebSocket connection to 'ws://wails.localhost:34115' failed` error. It is Vite's HMR socket being refused by App Transport Security, and it falls back on its own.
