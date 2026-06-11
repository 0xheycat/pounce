# Brand assets

Generated, reusable brand art for Pounce. Safe to regenerate; see commands below.

| File | Size | Use |
| --- | --- | --- |
| `hero.png` | 1280×448 | Top banner in the root `README.md`. |
| `social-preview.png` | 1280×640 | GitHub repo **social preview** + link unfurls (Twitter/Discord/Slack). |
| `../../dashboard/public/pounce.svg` | vector | Favicon + base logo (the orb). |
| `../../dashboard/public/pounce-192.png` / `pounce-512.png` | 192/512 | PWA / app icons. |

## Set the GitHub social preview

GitHub does **not** read an image from the repo automatically. Upload it once:

1. Repo **Settings** → **General** → scroll to **Social preview**.
2. **Edit** → upload `docs/assets/social-preview.png` (1280×640, < 1 MB).

Now every link to the repo unfurls with the Pounce card.

## Regenerate (ImageMagick)

These were produced with ImageMagick + DejaVu Sans. To rebuild after a brand tweak, re-run the `magick` commands used to create them (gradient background + composited orb logo + text). Keep the palette consistent:

- Background gradient: `#241a4d`/`#2a1f57` → `#070711`
- Accent: `#7c5cff`  ·  Core/secondary text: `#c4b5fd`  ·  Muted text: `#9aa0b5`

## Note

`hero.png` and `social-preview.png` are **brand art**, not product screenshots. Real UI captures (`dashboard.png`, `demo.gif`, …) belong in [`../screenshots/`](../screenshots/).
