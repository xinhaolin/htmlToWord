# Launch operations

The [publishing kit](publishing-kit.md) contains accurate community copy and a first-week measurement plan. The English and Chinese technical articles are original project implementation notes; generated HTML lives under docs/articles.

Rebuild the static site with `python3 scripts/build-docs.py`, then run `python3 scripts/build-docs.py --check`. The site uses no third-party JavaScript, trackers, external fonts or hosted conversion API. GitHub Pages publishes main:/docs. Canonicals, hreflang, Open Graph tags, JSON-LD, Atom feed and sitemap use https://xinhaolin.github.io/htmlToWord/.

Media provenance: docs/assets/manifest.json records the original DOCX screenshots. docs/assets/editing-demo.json records the edited office-client capture sequence, its checks and limitations. The social preview is a composition of project typography and the real report render.
